package khl

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// Gateway returns the url for websocket gateway.
// FYI: https://developer.kaiheila.cn/doc/http/gateway#%E8%8E%B7%E5%8F%96%E7%BD%91%E5%85%B3%E8%BF%9E%E6%8E%A5%E5%9C%B0%E5%9D%80
func (s *Session) Gateway() (gateway string, err error) {
	u, _ := url.Parse(EndpointGatewayIndex)
	q := u.Query()
	q.Set("compress", "0")
	if s.Identify.Compress {
		q.Set("compress", "1")
	}
	u.RawQuery = q.Encode()
	response, err := s.Request("GET", u.String(), nil)
	if err != nil {
		return
	}

	temp := struct {
		URL string `json:"url"`
	}{}

	err = json.Unmarshal(response, &temp)
	if err != nil {
		return
	}
	gateway = temp.URL
	return
}

// MessageListOption is the type for optional arguments for MessageList request.
type MessageListOption func(values url.Values)

// MessageListWithMsgID adds optional `msg_id` argument to MessageList request.
func MessageListWithMsgID(msgID string) MessageListOption {
	return func(values url.Values) {
		values.Set("msg_id", msgID)
	}
}

// MessageListWithPin adds optional `pin` argument to MessageList request.
func MessageListWithPin(pin bool) MessageListOption {
	return func(values url.Values) {
		if pin {
			values.Set("pin", "1")
		} else {
			values.Set("pin", "0")
		}
	}
}

// MessageListFlag is the type for the flag of MessageList.
type MessageListFlag string

// These are the usable flags
const (
	MessageListFlagBefore MessageListFlag = "before"
	MessageListFlagAround MessageListFlag = "around"
	MessageListFlagAfter  MessageListFlag = "after"
)

// MessageListWithFlag adds optional `flag` argument to MessageList request.
func MessageListWithFlag(flag MessageListFlag) MessageListOption {
	return func(values url.Values) {
		values.Set("flag", string(flag))
	}
}

// MessageList returns a list of messages of a channel.
// FYI: https://developer.kaiheila.cn/doc/http/message#%E8%8E%B7%E5%8F%96%E9%A2%91%E9%81%93%E8%81%8A%E5%A4%A9%E6%B6%88%E6%81%AF%E5%88%97%E8%A1%A8
func (s *Session) MessageList(targetID string, options ...MessageListOption) (ms []*DetailedChannelMessage, err error) {
	var response []byte
	u, _ := url.Parse(EndpointMessageList)
	q := u.Query()
	q.Set("target_id", targetID)
	for _, item := range options {
		item(q)
	}
	u.RawQuery = q.Encode()
	response, err = s.Request("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, &ms)
	if err != nil {
		return nil, err
	}
	return ms, nil
}

// MessageCreateBase is the common arguments for message creation.
type MessageCreateBase struct {
	Type     MessageType `json:"type,omitempty"`
	TargetID string      `json:"target_id,omitempty"`
	Content  string      `json:"content,omitempty"`
	Quote    string      `json:"quote,omitempty"`
	Nonce    string      `json:"nonce,omitempty"`
}

// MessageCreate is the type for message creation arguments.
type MessageCreate struct {
	MessageCreateBase
	TempTargetID string `json:"temp_target_id,omitempty"`
}

// MessageResp is the type for response for MessageCreate.
type MessageResp struct {
	MsgID        string         `json:"msg_id"`
	MegTimestamp MilliTimeStamp `json:"meg_timestamp"`
	Nonce        string         `json:"nonce"`
}

// MessageCreate creates a message.
// FYI: https://developer.kaiheila.cn/doc/http/message#%E5%8F%91%E9%80%81%E9%A2%91%E9%81%93%E8%81%8A%E5%A4%A9%E6%B6%88%E6%81%AF
func (s *Session) MessageCreate(m *MessageCreate) (resp *MessageResp, err error) {
	var response []byte
	response, err = s.Request("POST", EndpointMessageCreate, m)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, resp)
	if err != nil {
		return nil, err
	}
	return
}

// MessageUpdateBase is the shared arguments for message update related requests.
type MessageUpdateBase struct {
	MsgID   string `json:"msg_id"`
	Content string `json:"content"`
	Quote   string `json:"quote,omitempty"`
}

// MessageUpdate is the request data for MessageUpdate.
type MessageUpdate struct {
	MessageCreateBase
	TempTargetID string `json:"temp_target_id,omitempty"`
}

// MessageUpdate updates a message.
// FYI: https://developer.kaiheila.cn/doc/http/message#%E6%9B%B4%E6%96%B0%E9%A2%91%E9%81%93%E8%81%8A%E5%A4%A9%E6%B6%88%E6%81%AF
func (s *Session) MessageUpdate(m *MessageUpdate) (err error) {
	_, err = s.Request("POST", EndpointMessageUpdate, m)
	return
}

// MessageDelete deletes a message.
// FYI: https://developer.kaiheila.cn/doc/http/message#%E5%88%A0%E9%99%A4%E9%A2%91%E9%81%93%E8%81%8A%E5%A4%A9%E6%B6%88%E6%81%AF
func (s *Session) MessageDelete(msgID string) (err error) {
	_, err = s.Request("POST", EndpointMessageDelete, struct {
		MsgID string `json:"msg_id"`
	}{msgID})
	return
}

// ReactedUser is the type for every user reacted to a specific message with a specific emoji.
type ReactedUser struct {
	User
	ReactionTime MilliTimeStamp `json:"reaction_time"`
	TagInfo      struct {
		Color string `json:"color"`
		Text  string `json:"text"`
	} `json:"tag_info"`
}

// MessageReactionList returns the list of the reacted users with a specific emoji to a message.
// FYI: https://developer.kaiheila.cn/doc/http/message#%E8%8E%B7%E5%8F%96%E9%A2%91%E9%81%93%E6%B6%88%E6%81%AF%E6%9F%90%E5%9B%9E%E5%BA%94%E7%9A%84%E7%94%A8%E6%88%B7%E5%88%97%E8%A1%A8
func (s *Session) MessageReactionList(msgID, emoji string) (us []*ReactedUser, err error) {
	u, _ := url.Parse(EndpointMessageReactionList)
	q := u.Query()
	q.Add("msg_id", msgID)
	q.Add("emoji", emoji)
	u.RawQuery = q.Encode()
	var response []byte
	response, err = s.Request("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, &us)
	if err != nil {
		return nil, err
	}
	return us, nil
}

// MessageAddReaction add a reaction to a message as the bot.
// FYI: https://developer.kaiheila.cn/doc/http/message#%E7%BB%99%E6%9F%90%E4%B8%AA%E6%B6%88%E6%81%AF%E6%B7%BB%E5%8A%A0%E5%9B%9E%E5%BA%94
func (s *Session) MessageAddReaction(msgID, emoji string) (err error) {
	_, err = s.Request("POST", EndpointMessageAddReaction, struct {
		MsgID string `json:"msg_id"`
		Emoji string `json:"emoji"`
	}{msgID, emoji})
	return err
}

// MessageDeleteReaction deletes a reaction of a user from a message.
// FYI: https://developer.kaiheila.cn/doc/http/message#%E5%88%A0%E9%99%A4%E6%B6%88%E6%81%AF%E7%9A%84%E6%9F%90%E4%B8%AA%E5%9B%9E%E5%BA%94
func (s *Session) MessageDeleteReaction(msgID, emoji string, userID string) (err error) {
	_, err = s.Request("POST", EndpointMessageDeleteReaction, struct {
		MsgID  string `json:"msg_id"`
		Emoji  string `json:"emoji"`
		UserID string `json:"user_id,omitempty"`
	}{msgID, emoji, userID})
	return err
}

// ChannelList lists all channels from a guild.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E8%8E%B7%E5%8F%96%E9%A2%91%E9%81%93%E5%88%97%E8%A1%A8
func (s *Session) ChannelList(guildID string, page *PageSetting) (cs []*Channel, meta *PageInfo, err error) {
	var response []byte
	u, _ := url.Parse(EndpointChannelList)
	q := u.Query()
	q.Set("guild_id", guildID)
	u.RawQuery = q.Encode()
	response, meta, err = s.RequestWithPage("GET", u.String(), page)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(response, &cs)
	if err != nil {
		return nil, nil, err
	}
	return cs, meta, nil
}

// ChannelView returns the detailed information for a channel.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E8%8E%B7%E5%8F%96%E9%A2%91%E9%81%93%E8%AF%A6%E6%83%85
func (s *Session) ChannelView(channelID string) (c *Channel, err error) {
	var response []byte
	u, _ := url.Parse(EndpointChannelView)
	q := u.Query()
	q.Set("target_id", channelID)
	u.RawQuery = q.Encode()
	response, err = s.Request("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ChannelCreate is the arguments for creating a channel.
type ChannelCreate struct {
	GuildID      string      `json:"guild_id"`
	ParentID     string      `json:"parent_id,omitempty"`
	Name         string      `json:"name"`
	Type         ChannelType `json:"type,omitempty"`
	LimitAmount  int         `json:"limit_amount,omitempty"`
	VoiceQuality int         `json:"voice_quality,omitempty"`
}

// ChannelCreate creates a channel.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E5%88%9B%E5%BB%BA%E9%A2%91%E9%81%93
func (s *Session) ChannelCreate(cc *ChannelCreate) (c *Channel, err error) {
	var response []byte
	response, err = s.Request("POST", EndpointChannelCreate, cc)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, c)
	if err != nil {
		return nil, err
	}
	return c, err
}

// ChannelDelete deletes a channel.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E5%88%A0%E9%99%A4%E9%A2%91%E9%81%93
func (s *Session) ChannelDelete(channelID string) (err error) {
	_, err = s.Request("POST", EndpointChannelRoleDelete, struct {
		ChannelID string `json:"channel_id"`
	}{channelID})
	return err
}

// ChannelMoveUsers moves users to a channel.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E8%AF%AD%E9%9F%B3%E9%A2%91%E9%81%93%E4%B9%8B%E9%97%B4%E7%A7%BB%E5%8A%A8%E7%94%A8%E6%88%B7
func (s *Session) ChannelMoveUsers(targetChannelID string, userIDs []string) (err error) {
	_, err = s.Request("POST", EndpointChannelMoveUser, struct {
		TargetID string   `json:"target_id"`
		UserIDs  []string `json:"user_ids"`
	}{targetChannelID, userIDs})
	return err
}

// ChannelRoleIndex is the role and permission list of a channel.
type ChannelRoleIndex struct {
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites"`
	PermissionUsers      []struct {
		User  User           `json:"user"`
		Allow RolePermission `json:"allow"`
		Deny  RolePermission `json:"deny"`
	} `json:"permission_users"`
	PermissionSync bool `json:"permission_sync"`
}

// ChannelRoleIndex returns the role and permission list of the channel.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E9%A2%91%E9%81%93%E8%A7%92%E8%89%B2%E6%9D%83%E9%99%90%E8%AF%A6%E6%83%85
func (s *Session) ChannelRoleIndex(channelID string) (cr *ChannelRoleIndex, err error) {
	var response []byte
	u, _ := url.Parse(EndpointChannelRoleIndex)
	q := u.Query()
	q.Set("channel_id", channelID)
	u.RawQuery = q.Encode()
	response, err = s.Request("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, cr)
	if err != nil {
		return nil, err
	}
	return cr, err
}

// ChannelRoleBase is the common arguments for channel role requests.
type ChannelRoleBase struct {
	ChannelID string `json:"channel_id"`
	Type      string `json:"type,omitempty"`
	Value     string `json:"value,omitempty"`
}

// ChannelRoleCreate is the request query data for ChannelRoleCreate.
type ChannelRoleCreate ChannelRoleBase

// ChannelRoleCreate creates a role for a channel
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E5%88%9B%E5%BB%BA%E9%A2%91%E9%81%93%E8%A7%92%E8%89%B2%E6%9D%83%E9%99%90
func (s *Session) ChannelRoleCreate(crc *ChannelRoleCreate) (err error) {
	_, err = s.Request("POST", EndpointChannelRoleCreate, crc)
	return err
}

// ChannelRoleUpdate is the request query data for ChannelRoleUpdate
type ChannelRoleUpdate struct {
	ChannelRoleBase
	Allow RolePermission `json:"allow,omitempty"`
	Deny  RolePermission `json:"deny,omitempty"`
}

// ChannelRoleUpdateResp is the response of ChannelRoleUpdate
type ChannelRoleUpdateResp struct {
	UserID string         `json:"user_id"`
	RoleID string         `json:"role_id"`
	Allow  RolePermission `json:"allow"`
	Deny   RolePermission `json:"deny"`
}

// ChannelRoleUpdate updates a role from channel setting.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E6%9B%B4%E6%96%B0%E9%A2%91%E9%81%93%E8%A7%92%E8%89%B2%E6%9D%83%E9%99%90
func (s *Session) ChannelRoleUpdate(cru *ChannelRoleUpdate) (crur *ChannelRoleUpdateResp, err error) {
	var response []byte
	response, err = s.Request("POST", EndpointChannelRoleUpdate, cru)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, crur)
	if err != nil {
		return nil, err
	}
	return crur, nil
}

// ChannelRoleDelete is the type for settings when deleting a role from channel setting.
type ChannelRoleDelete ChannelRoleBase

// ChannelRoleDelete deletes a role form channel setting.
// FYI: https://developer.kaiheila.cn/doc/http/channel#%E5%88%A0%E9%99%A4%E9%A2%91%E9%81%93%E8%A7%92%E8%89%B2%E6%9D%83%E9%99%90
func (s *Session) ChannelRoleDelete(crd *ChannelRoleDelete) (err error) {
	_, err = s.Request("POST", EndpointChannelRoleDelete, crd)
	return err
}

// UserChatCreate creates a direct chat session.
// FYI: https://developer.kaiheila.cn/doc/http/user-chat#%E5%88%9B%E5%BB%BA%E7%A7%81%E4%BF%A1%E8%81%8A%E5%A4%A9%E4%BC%9A%E8%AF%9D
func (s *Session) UserChatCreate(UserID string) (uc *UserChat, err error) {
	var response []byte
	response, err = s.Request("POST", EndpointUserChatCreate, struct {
		TargetID string `json:"target_id"`
	}{UserID})
	if err != nil {
		return nil, err
	}
	uc = &UserChat{}
	err = json.Unmarshal(response, uc)
	if err != nil {
		return nil, err
	}
	return uc, err
}

// UserChatDelete deletes a direct chat session.
// FYI: https://developer.kaiheila.cn/doc/http/user-chat#%E5%88%9B%E5%BB%BA%E7%A7%81%E4%BF%A1%E8%81%8A%E5%A4%A9%E4%BC%9A%E8%AF%9D
func (s *Session) UserChatDelete(ChatCode string) (err error) {
	_, err = s.Request("POST", EndpointUserChatDelete, struct {
		ChatCode string `json:"chat_code"`
	}{ChatCode: ChatCode})
	return err
}

// DirectMessageCreate is the struct for settings of creating a message in direct chat.
type DirectMessageCreate struct {
	MessageCreateBase
	ChatCode string `json:"chat_code,omitempty"`
}

// DirectMessageCreate creates a message in direct chat.
// FYI: https://developer.kaiheila.cn/doc/http/direct-message#%E5%8F%91%E9%80%81%E7%A7%81%E4%BF%A1%E8%81%8A%E5%A4%A9%E6%B6%88%E6%81%AF
func (s *Session) DirectMessageCreate(create *DirectMessageCreate) (mr *MessageResp, err error) {
	var response []byte
	response, err = s.Request("POST", EndpointDirectMessageCreate, create)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, mr)
	if err != nil {
		return nil, err
	}
	return mr, nil
}

// DirectMessageUpdate is the type for settings of updating a message in direct chat.
type DirectMessageUpdate MessageUpdateBase

// DirectMessageUpdate updates a message in direct chat.
// FYI: https://developer.kaiheila.cn/doc/http/direct-message#%E6%9B%B4%E6%96%B0%E7%A7%81%E4%BF%A1%E8%81%8A%E5%A4%A9%E6%B6%88%E6%81%AF
func (s *Session) DirectMessageUpdate(update *DirectMessageUpdate) (err error) {
	_, err = s.Request("POST", EndpointDirectMessageUpdate, update)
	return err
}

// DirectMessageDelete deletes a message in direct chat.
// FYI: https://developer.kaiheila.cn/doc/http/direct-message#%E5%88%A0%E9%99%A4%E7%A7%81%E4%BF%A1%E8%81%8A%E5%A4%A9%E6%B6%88%E6%81%AF
func (s *Session) DirectMessageDelete(msgID string) (err error) {
	_, err = s.Request("POST", EndpointDirectMessageDelete, struct {
		MsgID string `json:"msg_id"`
	}{msgID})
	return err
}

// GuildList returns a list of guild that bot joins.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E8%8E%B7%E5%8F%96%E5%BD%93%E5%89%8D%E7%94%A8%E6%88%B7%E5%8A%A0%E5%85%A5%E7%9A%84%E6%9C%8D%E5%8A%A1%E5%99%A8%E5%88%97%E8%A1%A8
func (s *Session) GuildList(page *PageSetting) (gs []*Guild, meta *PageInfo, err error) {
	var response []byte
	response, meta, err = s.RequestWithPage("GET", EndpointGuildList, page)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(response, &gs)
	if err != nil {
		return nil, nil, err
	}
	return gs, meta, nil
}

// GuildView returns a detailed info for a guild.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E8%8E%B7%E5%8F%96%E6%9C%8D%E5%8A%A1%E5%99%A8%E8%AF%A6%E6%83%85
func (s *Session) GuildView(guildID string) (g *Guild, err error) {
	var response []byte
	u, _ := url.Parse(EndpointGuildView)
	q := u.Query()
	q.Add("guild_id", guildID)
	u.RawQuery = q.Encode()
	response, err = s.Request("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	g = &Guild{}
	err = json.Unmarshal(response, g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// GuildUserListOption is the type for optional arguments for GuildUserList request.
type GuildUserListOption func(values url.Values)

// GuildUserListWithChannelID adds optional `channel_id` argument to GuildUserList request.
func GuildUserListWithChannelID(id string) GuildUserListOption {
	return func(values url.Values) {
		values.Set("channel_id", id)
	}
}

// GuildUserListWithSearch adds optional `search` argument to GuildUserList request.
func GuildUserListWithSearch(search string) GuildUserListOption {
	return func(values url.Values) {
		values.Set("search", search)
	}
}

// GuildUserListWithRoleID adds optional `role_id` argument to GuildUserList request.
func GuildUserListWithRoleID(roleID int64) GuildUserListOption {
	return func(values url.Values) {
		values.Set("role_id", strconv.FormatInt(roleID, 10))
	}
}

// GuildUserListWithMobileVerified adds optional `mobile_verified` argument to GuildUserList request.
func GuildUserListWithMobileVerified(verified bool) GuildUserListOption {
	return func(values url.Values) {
		if verified {
			values.Set("mobile_verified", "1")
		} else {
			values.Set("mobile_verified", "0")
		}
	}
}

// GuildUserListWithActiveTime adds optional `active_time` argument to GuildUserList request.
func GuildUserListWithActiveTime(activeTime bool) GuildUserListOption {
	return func(values url.Values) {
		if activeTime {
			values.Set("active_time", "1")
		} else {
			values.Set("active_time", "0")
		}
	}
}

// GuildUserListWithJoinedAt adds optional `joined_at` argument to GuildUserList request.
func GuildUserListWithJoinedAt(joinedAt bool) GuildUserListOption {
	return func(values url.Values) {
		if joinedAt {
			values.Set("joined_at", "1")
		} else {
			values.Set("joined_at", "0")
		}
	}
}

// GuildUserList returns the list of users in a guild.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E8%8E%B7%E5%8F%96%E6%9C%8D%E5%8A%A1%E5%99%A8%E4%B8%AD%E7%9A%84%E7%94%A8%E6%88%B7%E5%88%97%E8%A1%A8
func (s *Session) GuildUserList(guildID string, page *PageSetting, options ...GuildUserListOption) (us []*User, meta *PageInfo, err error) {
	var response []byte
	p := &PageInfo{}
	u, _ := url.Parse(EndpointGuildUserList)
	q := u.Query()
	q.Set("guild_id", guildID)
	for _, item := range options {
		item(q)
	}
	u.RawQuery = q.Encode()
	response, p, err = s.RequestWithPage("GET", u.String(), page)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(response, &us)
	if err != nil {
		return nil, nil, err
	}
	return us, p, err
}

// GuildNickname is the arguments for GuildNickname.
type GuildNickname struct {
	GuildID  string `json:"guild_id"`
	Nickname string `json:"nickname,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// GuildNickname changes the nickname of a user in a guild.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E4%BF%AE%E6%94%B9%E6%9C%8D%E5%8A%A1%E5%99%A8%E4%B8%AD%E7%94%A8%E6%88%B7%E7%9A%84%E6%98%B5%E7%A7%B0
func (s *Session) GuildNickname(gn *GuildNickname) (err error) {
	_, err = s.Request("POST", EndpointGuildNickName, gn)
	return err
}

// GuildLeave let the bot leave a guild.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E7%A6%BB%E5%BC%80%E6%9C%8D%E5%8A%A1%E5%99%A8
func (s *Session) GuildLeave(guildID string) (err error) {
	_, err = s.Request("POST", EndpointGuildLeave, struct {
		GuildID string `json:"guild_id"`
	}{guildID})
	return err
}

// GuildKickout force a user to leave a guild.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E8%B8%A2%E5%87%BA%E6%9C%8D%E5%8A%A1%E5%99%A8
func (s *Session) GuildKickout(guildID, targetID string) (err error) {
	_, err = s.Request("POST", EndpointGuildKickout, struct {
		GuildID  string `json:"guild_id"`
		TargetID string `json:"target_id"`
	}{guildID, targetID})
	return err
}

// GuildMuteList is the type for users that got muted in a guild.
type GuildMuteList struct {
	Mic     []string `json:"1"`
	Headset []string `json:"2"`
}

// GuildMuteList returns the list of users got mutes in mic or earphone.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E6%9C%8D%E5%8A%A1%E5%99%A8%E9%9D%99%E9%9F%B3%E9%97%AD%E9%BA%A6%E5%88%97%E8%A1%A8
func (s *Session) GuildMuteList(guildID string) (gml *GuildMuteList, err error) {
	var response []byte
	u, _ := url.Parse(EndpointGuildMuteList)
	q := u.Query()
	q.Set("guild_id", guildID)
	u.RawQuery = q.Encode()
	response, err = s.Request("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	gml = &GuildMuteList{}
	err = json.Unmarshal(response, gml)
	if err != nil {
		return nil, err
	}
	return gml, nil
}

// MuteType is the type for mute status.
type MuteType int8

// These are all mute types.
const (
	MuteTypeMic MuteType = iota + 1
	MuteTypeHeadset
)

// GuildMuteSetting is the type for arguments of GuildMuteSetting.
type GuildMuteSetting struct {
	GuildID string   `json:"guild_id"`
	UserID  string   `json:"user_id"`
	Type    MuteType `json:"type"`
}

// GuildMuteCreate revokes a users privilege of using mic or headset.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E6%B7%BB%E5%8A%A0%E6%9C%8D%E5%8A%A1%E5%99%A8%E9%9D%99%E9%9F%B3%E6%88%96%E9%97%AD%E9%BA%A6
func (s *Session) GuildMuteCreate(gms *GuildMuteSetting) (err error) {
	_, err = s.Request("POST", EndpointGuildMuteCreate, gms)
	return err
}

// GuildMuteDelete re-grants a users privilege of using mic or headset.
// FYI: https://developer.kaiheila.cn/doc/http/guild#%E5%88%A0%E9%99%A4%E6%9C%8D%E5%8A%A1%E5%99%A8%E9%9D%99%E9%9F%B3%E6%88%96%E9%97%AD%E9%BA%A6
func (s *Session) GuildMuteDelete(gms *GuildMuteSetting) (err error) {
	_, err = s.Request("POST", EndpointGuildMuteDelete, gms)
	return err
}

// GuildRoleList returns the roles in a guild.
// FYI: https://developer.kaiheila.cn/doc/http/guild-role#%E8%8E%B7%E5%8F%96%E6%9C%8D%E5%8A%A1%E5%99%A8%E8%A7%92%E8%89%B2%E5%88%97%E8%A1%A8
func (s *Session) GuildRoleList(guildID string, page *PageSetting) (rs []*Role, meta *PageInfo, err error) {
	var response []byte
	p := &PageInfo{}
	u, _ := url.Parse(EndpointGuildRoleList)
	q := u.Query()
	q.Add("guild_id", guildID)
	u.RawQuery = q.Encode()
	response, p, err = s.RequestWithPage("GET", u.String(), page)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(response, &rs)
	if err != nil {
		return nil, nil, err
	}
	return rs, p, err
}

// UserMe returns the bot info.
// FYI: https://developer.kaiheila.cn/doc/http/user
func (s *Session) UserMe() (u *User, err error) {
	var response []byte
	response, err = s.Request("GET", EndpointUserMe, nil)
	if err != nil {
		return nil, err
	}
	u = &User{}
	err = json.Unmarshal(response, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// RequestWithPage is the wrapper for internal list GET request, you would prefer to use other method other than this.
func (s *Session) RequestWithPage(method, u string, page *PageSetting) (response []byte, meta *PageInfo, err error) {
	ur, _ := url.Parse(u)
	q := ur.Query()
	if page.Page != nil {
		q.Add("page", strconv.Itoa(*page.Page))
	}
	if page.PageSize != nil {
		q.Add("page_size", strconv.Itoa(*page.PageSize))
	}
	if page.Sort != nil {
		q.Add("sort", *page.Sort)
	}
	ur.RawQuery = q.Encode()
	resp, err := s.Request(method, u, nil)
	if err != nil {
		return nil, nil, err
	}
	g := &GeneralListData{}
	err = json.Unmarshal(resp, g)
	if err != nil {
		return nil, nil, err
	}
	return g.Items, &g.Meta, err
}

// Request is the wrapper for internal request method, you would prefer to use other method other than this.
func (s *Session) Request(method, url string, data interface{}) (response []byte, err error) {
	return s.request(method, url, data, 0)
}

func (s *Session) request(method, url string, data interface{}, sequence int) (response []byte, err error) {
	var body []byte
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return
		}
	}
	//s.log(LogTrace, "Api Request %s %s\n", method, url)
	e := s.Logger.Trace().Str("method", method).Str("url", url)
	e = addCaller(e)
	if len(body) != 0 {
		e = e.Bytes("payload", body)
	}
	e.Msg("http api request")
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", s.Identify.Token)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	e = addCaller(s.Logger.Trace())
	for k, v := range req.Header {
		e = e.Strs(k, v)
		//s.log(LogTrace, "Api Request Header %s = %+v\n", k, v)
	}
	e.Msg("http api request headers")
	resp, err := s.Client.Do(req)
	if err != nil {
		addCaller(s.Logger.Error()).Err("err", err).Msg("")
		return
	}
	defer func() {
		err2 := resp.Body.Close()
		if err2 != nil {
			addCaller(s.Logger.Error()).Msg("error closing resp body")
			//s.log(LogError, "error closing resp body")
		}
	}()

	var respByte []byte

	respByte, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		addCaller(s.Logger.Error()).Err("err", err).Msg("")
		return
	}
	addCaller(s.Logger.Trace()).Int("status_code", resp.StatusCode).
		Str("status", resp.Status).
		Bytes("body", respByte).
		Msg("http response")
	//s.log(LogTrace, "Api Response Status %s\n", resp.Status)
	e = s.Logger.Trace()
	e = addCaller(e)
	for k, v := range resp.Header {
		e = e.Strs(k, v)
		//s.log(LogTrace, "Api Response Header %s = %+v\n", k, v)
	}
	e.Msg("http response headers")
	//s.log(LogTrace, "Api Response Body %s", respByte)
	var r EndpointGeneralResponse
	err = json.Unmarshal(respByte, &r)
	if err != nil {
		addCaller(s.Logger.Error()).Err("err", err).Msg("response unmarshal error")
		//s.log(LogError, "Api Response Unmarshal Error %s", err)
		return
	}
	if r.Code != 0 {
		addCaller(s.Logger.Error()).Int("code", r.Code).Str("error_msg", r.Message).Msg("api response error")
		//s.log(LogError, "Api Response Error Code %d, Message %s", r.Code, r.Message)
		return
	}
	response = r.Data
	return
}
