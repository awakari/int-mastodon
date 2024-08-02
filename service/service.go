package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ap "github.com/awakari/int-mastodon/api/grpc/int-activitypub"
	"github.com/awakari/int-mastodon/config"
	"github.com/awakari/int-mastodon/model"
	"github.com/awakari/int-mastodon/service/writer"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/r3labs/sse/v2"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	SearchAndAdd(ctx context.Context, interestId, groupId, q string, limit uint32, typ model.SearchType) (n uint32, err error)
	HandleLiveStream(ctx context.Context) (err error)
}

type mastodon struct {
	clientHttp     *http.Client
	userAgent      string
	cfg            config.MastodonConfig
	svcAp          ap.Service
	w              writer.Service
	typeCloudEvent string
}

const limitRespBodyLen = 1_048_576
const limitRespBodyLenErr = 1_024
const groupIdDefault = "default"
const tagNoBot = "#nobot"

func NewService(
	clientHttp *http.Client,
	userAgent string,
	cfgMastodon config.MastodonConfig,
	svcAp ap.Service,
	w writer.Service,
	typeCloudEvent string,
) Service {
	return mastodon{
		clientHttp:     clientHttp,
		userAgent:      userAgent,
		cfg:            cfgMastodon,
		svcAp:          svcAp,
		w:              w,
		typeCloudEvent: typeCloudEvent,
	}
}

func (m mastodon) SearchAndAdd(ctx context.Context, interestId, groupId, q string, limit uint32, typ model.SearchType) (n uint32, errs error) {
	for n < limit {
		reqQuery := "?q=" + url.QueryEscape(q) + "&type=" + typ.String() + "&resolve=true&offset=" + strconv.Itoa(int(n)) + "&limit=" + strconv.Itoa(int(limit-n))
		var req *http.Request
		var err error
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.Endpoint.Search+reqQuery, nil)
		var resp *http.Response
		if err == nil {
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Authorization", "Bearer "+m.cfg.Client.Token)
			req.Header.Add("User-Agent", m.userAgent)
			resp, err = m.clientHttp.Do(req)
		}
		var data []byte
		if err == nil && resp != nil {
			data, err = io.ReadAll(io.LimitReader(resp.Body, limitRespBodyLen))
			_ = resp.Body.Close()
		}
		var results model.Results
		if err == nil {
			err = json.Unmarshal(data, &results)
		}
		if err != nil {
			errs = errors.Join(errs, err)
			break
		}
		if typ == model.SearchTypeStatuses {
			countResults := len(results.Statuses)
			if countResults == 0 {
				break
			}
			n += uint32(countResults)
			for _, st := range results.Statuses {
				err = m.processFoundStatus(ctx, st, interestId, groupId, q)
				if err != nil {
					err = errors.Join(errs, err)
				}
			}
		} else if typ == model.SearchTypeAccounts {
			countResults := len(results.Accounts)
			if countResults == 0 {
				break
			}
			n += uint32(countResults)
			for _, acc := range results.Accounts {
				err = m.processFoundAccount(ctx, acc, interestId, groupId, q, false)
				if err != nil {
					errs = errors.Join(errs, err)
				}
			}
		}
	}
	return
}

func (m mastodon) processFoundStatus(ctx context.Context, s model.Status, interestId, groupId, q string) (err error) {
	if s.Sensitive {
		err = fmt.Errorf("found account %s skip due to sensitive flag", s.Account.Uri)
	}
	acc := s.Account
	if err == nil && acc.FollowersCount < m.cfg.CountMin.Followers {
		err = fmt.Errorf("found account %s skip due low followers count %d", acc.Uri, acc.FollowersCount)
	}
	if err == nil && acc.StatusesCount < m.cfg.CountMin.Followers {
		err = fmt.Errorf("found account %s skip due low post count %d", acc.Uri, acc.StatusesCount)
	}
	err = m.processFoundAccount(ctx, acc, interestId, groupId, q, true)
	return
}

func (m mastodon) processFoundAccount(ctx context.Context, acc model.Account, interestId, groupId, q string, delegateFollow bool) (err error) {
	if !acc.Discoverable {
		err = fmt.Errorf("found account %s skip due to no explicit discoverable flag set", acc.Uri)
	}
	if err == nil && acc.Indexable != nil && !*acc.Indexable {
		err = fmt.Errorf("found account %s skip due to no explicit indexable flag set", acc.Uri)
	}
	if err == nil && acc.Noindex {
		err = fmt.Errorf("found account %s skip due to noindex flag", acc.Uri)
	}
	if err == nil {
		for _, t := range acc.Tags {
			if strings.ToLower(t.Name) == tagNoBot {
				err = fmt.Errorf("found account %s skip due to %s tag", acc.Uri, tagNoBot)
				break
			}
		}
	}
	if err == nil {
		switch delegateFollow {
		case true:
			err = m.svcAp.Create(ctx, acc.Uri, groupId, "", interestId, q)
		default:
			err = m.follow(ctx, acc)
		}
	}
	return
}

func (m mastodon) follow(ctx context.Context, acc model.Account) (err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, m.cfg.Endpoint.Accounts+"/"+acc.Id+"/follow", nil)
	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+m.cfg.Client.Token)
		req.Header.Add("User-Agent", m.userAgent)
		resp, err = m.clientHttp.Do(req)
	}
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(io.LimitReader(resp.Body, limitRespBodyLenErr))
			err = fmt.Errorf(
				"failed to follow the account %s: request_url=%s, request_headers=%+v, response=%d/%s",
				acc.Acct,
				req.URL,
				req.Header,
				resp.StatusCode,
				string(data),
			)
		}
	}
	return
}

func (m mastodon) HandleLiveStream(ctx context.Context) (err error) {
	clientSse := sse.NewClient(m.cfg.Endpoint.Stream)
	clientSse.Headers["Authorization"] = "Bearer " + m.cfg.Client.Token
	clientSse.Headers["User-Agent"] = m.userAgent
	chSsEvts := make(chan *sse.Event)
	err = clientSse.SubscribeChanWithContext(ctx, "", chSsEvts)
	if err == nil {
		defer clientSse.Unsubscribe(chSsEvts)
		for {
			select {
			case ssEvt := <-chSsEvts:
				m.handleLiveStreamEvent(ctx, ssEvt)
			case <-ctx.Done():
				err = ctx.Err()
			case <-time.After(m.cfg.StreamTimeoutMax):
				err = fmt.Errorf("timeout while wating for new stream status")
			}
			if err != nil {
				break
			}
		}
	}
	return
}

func (m mastodon) handleLiveStreamEvent(ctx context.Context, ssEvt *sse.Event) {
	if "update" == string(ssEvt.Event) {

		var st model.Status
		err := json.Unmarshal(ssEvt.Data, &st)
		if err != nil {
			fmt.Printf("failed to unmarshal the live stream event data: %s\nerror: %s\n", string(ssEvt.Data), err)
		}

		// do not proceed if either of below conditions is true
		if st.Sensitive {
			return
		}
		if st.Visibility != "public" {
			return
		}
		acc := st.Account
		if !acc.Discoverable {
			return
		}
		if acc.Noindex {
			return
		}
		for _, t := range st.Tags {
			if strings.ToLower(t.Name) == tagNoBot {
				return
			}
		}
		for _, t := range acc.Tags {
			if strings.ToLower(t.Name) == tagNoBot {
				return
			}
		}

		if acc.FollowersCount < m.cfg.CountMin.Followers {
			return
		}
		if acc.StatusesCount < m.cfg.CountMin.Posts {
			return
		}

		addr := acc.Uri
		if addr == "" {
			addr = acc.Url
		}
		switch {
		case acc.Locked:
			// able to accept the follow request manually
			if addr == "" {
				addr = acc.Acct
			}
			_ = m.svcAp.Create(ctx, addr, groupIdDefault, addr, "", "")
		case acc.Indexable == nil || *acc.Indexable == true:
			// account allows explicitly to consume their posts
			evtAwk := m.convertStatus(st, addr)
			err = m.w.Write(context.TODO(), evtAwk, groupIdDefault, addr)
			if err != nil {
				fmt.Printf("failed to submit the live stream event, id=%s, src=%s, err=%s\n", evtAwk.Id, addr, err)
			}
		}
	}
	return
}

func (m mastodon) convertStatus(st model.Status, src string) (evtAwk *pb.CloudEvent) {
	evtAwk = &pb.CloudEvent{
		Id:          ksuid.New().String(),
		Source:      src,
		SpecVersion: model.CeSpecVersion,
		Type:        m.typeCloudEvent,
		Attributes: map[string]*pb.CloudEventAttributeValue{
			model.CeKeySubject: {
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: st.Account.DisplayName,
				},
			},
			model.CeKeyTime: {
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: timestamppb.New(st.CreatedAt.UTC()),
				},
			},
		},
		Data: &pb.CloudEvent_TextData{
			TextData: st.Content,
		},
	}
	if st.Language != "" {
		evtAwk.Attributes["language"] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: st.Language,
			},
		}
	}
	if st.Url != "" {
		evtAwk.Attributes[model.CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: st.Url,
			},
		}
	}
	var cats []string
	for _, t := range st.Tags {
		if t.Name != "" {
			cats = append(cats, t.Name)
		}
	}
	if len(cats) > 0 {
		evtAwk.Attributes[model.CeKeyCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(cats, " "),
			},
		}
	}
	if len(st.MediaAttachments) > 0 {
		att := st.MediaAttachments[0]
		evtAwk.Attributes[model.CeKeyAttachmentType] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: att.Type,
			},
		}
		u := att.PreviewUrl
		if u == "" {
			u = att.Url
		}
		evtAwk.Attributes[model.CeKeyAttachmentUrl] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: u,
			},
		}
	}
	return
}
