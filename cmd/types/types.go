package types

import (
	"bytes"
	"fmt"
	"io"
)

var participantId = 0

type Participant struct {
	Id     int
	Name   string
	Gender string
}

func NewParticipant(id int, name string, gender string) Participant {
	return Participant{
		Id:     id,
		Name:   name,
		Gender: gender,
	}
}

type Participants []Participant

func (p *Participants) GetParticipantById(id int) *Participant {
	for _, p := range *p {
		if p.Id == id {
			return &p
		}
	}
	return nil
}

func NewParticipants() Participants {
	return Participants{}
}

var talkTypeId = 0

type TalkType struct {
	Id         int
	Name       string
	MaxReplies int
	Color      string
}

func NewTalkType(name string, replies int, color string) TalkType {
	talkTypeId++
	return TalkType{
		Id:         talkTypeId,
		Name:       name,
		MaxReplies: replies,
		Color:      color,
	}
}

type TalkTypes []TalkType

func (t *TalkTypes) GetTalkTypeById(id int) *TalkType {
	for _, t := range *t {
		if t.Id == id {
			return &t
		}
	}
	return nil
}

func NewTalkTypes() TalkTypes {
	return TalkTypes{}
}

type Talk struct {
	Id          int
	Type        int
	Participant int
	ReplyTo     *Talk
}

func NewTalk(talkId int, participantId int, talkTypeId int, replyTo *Talk) Talk {
	return Talk{
		Id:          talkId,
		Type:        talkTypeId,
		Participant: participantId,
		ReplyTo:     replyTo,
	}
}

type Talks []Talk

func NewTalks() Talks {
	return Talks{}
}

type Page struct {
	TalkTypes    TalkTypes
	Participants Participants
	TalkBlocks   TalkBlocks
}

func NewPage() Page {
	return Page{
		TalkTypes:    NewTalkTypes(),
		Participants: NewParticipants(),
		TalkBlocks:   NewTalkBlocks(),
	}
}

type Message struct {
	message string
}

func NewMessage(message string) Message {
	return Message{
		message: message,
	}
}

type TalkBlock struct {
	Talk        Talk
	Participant Participant
	TalkType    TalkType
}

type TalkBlocks []TalkBlock

func NewTalkBlocks() TalkBlocks {
	return TalkBlocks{}
}

func NewTalkBlock(talk Talk, participant Participant, talk_type TalkType) TalkBlock {
	return TalkBlock{
		Talk:        talk,
		Participant: participant,
		TalkType:    talk_type,
	}
}

// Event represents Server-Sent Event.
// SSE explanation: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
type Event struct {
	// ID is used to set the EventSource object's last event ID value.
	ID []byte
	// Data field is for the message. When the EventSource receives multiple consecutive lines
	// that begin with data:, it concatenates them, inserting a newline character between each one.
	// Trailing newlines are removed.
	Data []byte
	// Event is a string identifying the type of event described. If this is specified, an event
	// will be dispatched on the browser to the listener for the specified event name; the website
	// source code should use addEventListener() to listen for named events. The onmessage handler
	// is called if no event name is specified for a message.
	Event []byte
	// Retry is the reconnection time. If the connection to the server is lost, the browser will
	// wait for the specified time before attempting to reconnect. This must be an integer, specifying
	// the reconnection time in milliseconds. If a non-integer value is specified, the field is ignored.
	Retry []byte
	// Comment line can be used to prevent connections from timing out; a server can send a comment
	// periodically to keep the connection alive.
	Comment []byte
}

// MarshalTo marshals Event to given Writer
func (ev *Event) MarshalTo(w io.Writer) error {
	// Marshalling part is taken from: https://github.com/r3labs/sse/blob/c6d5381ee3ca63828b321c16baa008fd6c0b4564/http.go#L16
	if len(ev.Data) == 0 && len(ev.Comment) == 0 {
		return nil
	}

	if len(ev.Data) > 0 {
		if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
			return err
		}

		sd := bytes.Split(ev.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}

		if len(ev.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
				return err
			}
		}

		if len(ev.Retry) > 0 {
			if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
				return err
			}
		}
	}

	if len(ev.Comment) > 0 {
		if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}

type TursoDB struct {
	Name          string   `json:"Name"`
	DbID          string   `json:"DbId"`
	Hostname      string   `json:"Hostname"`
	BlockReads    bool     `json:"block_reads"`
	BlockWrites   bool     `json:"block_writes"`
	AllowAttach   bool     `json:"allow_attach"`
	Regions       []string `json:"regions"`
	PrimaryRegion string   `json:"primaryRegion"`
	Type          string   `json:"type"`
	Version       string   `json:"version"`
	Group         string   `json:"group"`
	IsSchema      bool     `json:"is_schema"`
	Schema        string   `json:"schema"`
	Sleeping      bool     `json:"sleeping"`
}

type TursoDBs struct {
	Databases []TursoDB `json:"databases"`
}
