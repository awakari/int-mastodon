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
	"github.com/google/uuid"
	"github.com/r3labs/sse/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error)
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

func (m mastodon) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error) {
	var offset int
	for n < limit {
		reqQuery := "?q=" + url.QueryEscape(q) + "&type=statuses&offset=" + strconv.Itoa(offset) + "&limit=" + strconv.Itoa(int(limit))
		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.Endpoint.Search+reqQuery, nil)
		var resp *http.Response
		if err == nil {
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Authorization", "Bearer "+m.cfg.Client.Token)
			req.Header.Add("User-Agent", m.userAgent)
			resp, err = m.clientHttp.Do(req)
		}
		var data []byte
		if err == nil {
			data, err = io.ReadAll(io.LimitReader(resp.Body, limitRespBodyLen))
		}
		var results model.Results
		if err == nil {
			err = json.Unmarshal(data, &results)
		}
		if err == nil {
			countResults := len(results.Statuses)
			if countResults == 0 {
				break
			}
			offset += countResults
			for _, s := range results.Statuses {
				var errReqFollow error
				if s.Sensitive {
					errReqFollow = fmt.Errorf("found account %s skip due to sensitive flag", s.Account.Uri)
				}
				acc := s.Account
				if !acc.Discoverable {
					errReqFollow = fmt.Errorf("found account %s skip due to no explicit discoverable flag set", s.Account.Uri)
				}
				if acc.Indexable != nil && !*acc.Indexable {
					errReqFollow = fmt.Errorf("found account %s skip due to no explicit indexable flag set", s.Account.Uri)
				}
				if acc.Noindex {
					errReqFollow = fmt.Errorf("found account %s skip due to noindex flag", s.Account.Uri)
				}
				for _, t := range acc.Tags {
					if strings.ToLower(t.Name) == tagNoBot {
						errReqFollow = fmt.Errorf("found account %s skip due to %s tag", s.Account.Uri, tagNoBot)
						break
					}
				}
				if errReqFollow == nil && s.Account.FollowersCount < m.cfg.CountMin.Followers {
					errReqFollow = fmt.Errorf("found account %s skip due low followers count %d", s.Account.Uri, s.Account.FollowersCount)
				}
				if errReqFollow == nil && s.Account.StatusesCount < m.cfg.CountMin.Followers {
					errReqFollow = fmt.Errorf("found account %s skip due low post count %d", s.Account.Uri, s.Account.StatusesCount)
				}
				if errReqFollow == nil {
					errReqFollow = m.svcAp.Create(ctx, acc.Uri, groupId, "", subId, q)
				}
				if errReqFollow == nil {
					n++
				}
				if n > limit {
					break
				}
				err = errors.Join(err, errReqFollow)
			}
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
		Id:          uuid.NewString(),
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
