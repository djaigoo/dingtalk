package tools

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "strconv"
    "time"

    "github.com/djaigoo/httpclient"
    "github.com/pkg/errors"
)

// 钉钉文档：https://open.dingtalk.com/document/robots/custom-robot-access
// 每个机器人每分钟最多发送20条消息到群里，如果超过20条，会限流10分钟。

const dingTalkRequestURL = `https://oapi.dingtalk.com/robot/send`

type dingTalk struct {
    token  string
    secret string
}

func NewDingTalkBot(token, secret string) *dingTalk {
    return &dingTalk{
        token:  token,
        secret: secret,
    }
}

type ddMsg struct {
    Msgtype    string       `json:"msgtype"`
    Text       ddText       `json:"text"`
    Markdown   ddMarkdown   `json:"markdown"`
    ActionCard ddActionCard `json:"actionCard"`
    FeedCard   ddFeedCard   `json:"feedCard"`
    At         ddAt         `json:"at"`
}

type ddText struct {
    Content string `json:"content"`
}

type ddMarkdown struct {
    Title string `json:"title"`
    Text  string `json:"text"`
}

type ddActionCard struct {
    Title          string   `json:"title"`
    Text           string   `json:"text"`
    HideAvatar     string   `json:"hideAvatar"`
    BtnOrientation string   `json:"btnOrientation"`
    SingleTitle    string   `json:"singleTitle"`
    SingleURL      string   `json:"singleURL"`
    Btns           []ddBtns `json:"btns"`
}

type ddBtns struct {
    Title     string `json:"title"`
    ActionURL string `json:"actionURL"`
}

type ddLinks struct {
    Text       string `json:"text"`
    Title      string `json:"title"`
    PicURL     string `json:"picUrl"`
    MessageURL string `json:"messageUrl"`
}

type ddFeedCard struct {
    Links []ddLinks `json:"links"`
}

type ddAt struct {
    AtMobiles []string `json:"atMobiles"`
    IsAtAll   bool     `json:"isAtAll"`
}

type ddResponse struct {
    Errcode int    `json:"errcode"`
    Errmsg  string `json:"errmsg"`
}

type ddSendMsg struct {
    *dingTalk
    ddMsg
}

// Message
func (m *ddSendMsg) Message(text string) *ddSendMsg {
    m.ddMsg.Msgtype = "text"
    m.ddMsg.Text.Content = text
    return m
}

// Markdown
func (m *ddSendMsg) Markdown(title, text string) *ddSendMsg {
    m.ddMsg.Msgtype = "markdown"
    m.ddMsg.Markdown.Title = title
    m.ddMsg.Markdown.Text = text
    return m
}

// At
// atlist 传手机号
func (m *ddSendMsg) At(atall bool, atlist ...string) *ddSendMsg {
    m.ddMsg.At.IsAtAll = atall
    m.ddMsg.At.AtMobiles = atlist
    return m
}

// Do
func (m *ddSendMsg) Do() error {
    if m.ddMsg.Msgtype == "" {
        return errors.New("invalid msg type")
    }
    return m.dingTalk.send(m.ddMsg)
}

// Send 发送消息
func (dt *dingTalk) Send() *ddSendMsg {
    return &ddSendMsg{
        dingTalk: dt,
        ddMsg:    ddMsg{},
    }
}

func (dt *dingTalk) send(data interface{}) error {
    res := &ddResponse{}
    req := httpclient.Post(dingTalkRequestURL).AddQuery("access_token", dt.token)
    if dt.secret != "" {
        tn := time.Now().UnixMilli()
        ts := strconv.FormatInt(tn, 10)

        s := ts + "\n" + dt.secret
        h := hmac.New(sha256.New, []byte(dt.secret))
        h.Write([]byte(s))
        mac := h.Sum(nil)
        ms := base64.StdEncoding.EncodeToString(mac)

        req.AddQuery("timestamp", ts).AddQuery("sign", ms)
    }

    err := req.SetBody4Json(data).Do().ToJson(res)
    if err != nil {
        return err
    }
    if res.Errcode != 0 {
        return errors.New(res.Errmsg)
    }
    return nil
}
