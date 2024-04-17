package main

var participantId = 0

type Participant struct {
	Id   int
	Name string
}

func newParticipant(name string) Participant {
	participantId++
	return Participant{
		Id:   participantId,
		Name: name,
	}
}

type Participants []Participant

func (p *Participants) getParticipantById(id int) *Participant {
	for _, p := range *p {
		if p.Id == id {
			return &p
		}
	}
	return nil
}

func newParticipants() Participants {
	return Participants{
		newParticipant("Audun"),
		newParticipant("Mali"),
	}
}

var talkId = 0

type Talk struct {
	Id          int
	Type        string
	Participant int
}

func newTalk(participantId int) Talk {
	talkId++
	return Talk{
		Id:          talkId,
		Type:        "innlegg",
		Participant: participantId,
	}
}

type Talks []Talk

func newTalks() Talks {
	return Talks{}
}

type Page struct {
	Participants Participants
	TalkBlocks   TalkBlocks
}

func newPage() Page {
	return Page{
		Participants: newParticipants(),
		TalkBlocks:   newTalkBlocks(),
	}
}

type Message struct {
	message string
}

func newMessage(message string) Message {
	return Message{
		message: message,
	}
}

type TalkBlock struct {
	Talk        Talk
	Participant Participant
}

type TalkBlocks []TalkBlock

func newTalkBlocks() TalkBlocks {
	return TalkBlocks{}
}

func newTalkBlock(talk Talk, participant Participant) TalkBlock {
	return TalkBlock{
		Talk:        talk,
		Participant: participant,
	}
}
