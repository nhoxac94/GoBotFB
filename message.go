package main

type (
	Request struct {
		Object string `json:"object,omitempty"`
		Entry  []struct {
			Id        string      `json:"id,omitempty"`
			Time      int         `json:"time,omitempty"`
			Messaging []Messaging `json:"messaging,omitempty"`
		} `json:"entry"`
	}

	Messaging struct {
		Sender    *User    `json:"sender,omitempty"`
		Recipient *User    `json:"recipient,omitempty"`
		Timestamp int      `json:"timestamp,omitempty"`
		Message   *Message `json:"message,omitempty"`
	}

	User struct {
		Id string `json:"id,omitempty"`
	}

	Message struct {
		Mid        string      `json:"mid,omitempty"`
		Text       string      `json:"text,omitempty"`
		QuickReply *QuickReply `json:"quick_reply,omitempty"`
	}

	ResponseMessage struct {
		MessageType string      `json:"message_type"`
		Recipient   *User       `json:"recipient"`
		Message     *ResMessage `json:"message,omitempty"`
		Action      string      `json:"sender_action,omitempty"`
	}

	ResMessage struct {
		Text       string       `json:"text,omitempty"`
		QuickReply []QuickReply `json:"quick_replies,omitempty"`
	}

	QuickReply struct {
		ContentType string `json:"content_type,omitempty"`
		Title       string `json:"title,omitempty"`
		Payload     string `json:"payload"`
	}
)
