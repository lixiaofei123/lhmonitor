package notifier

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

func init() {

	fontContent, err := ioutil.ReadFile("./PingFang.ttc")
	if err == nil {
		font, err := truetype.Parse(fontContent)
		if err == nil {
			draw2d.RegisterFont(draw2d.FontData{Name: "PingFang"}, font)
		}
	}

	registerNotifier("qywx", NewQYWXNotifier())
}

type QYWXNotifier struct {
	key string
}

func NewQYWXNotifier() Notifier {
	return &QYWXNotifier{
		key: os.Getenv("QYWX_KEY"),
	}
}

type Image struct {
	Base64 string `json:"base64"`
	MD5    string `json:"md5"`
}

type QWMessage struct {
	MsgType string `json:"msgtype"`
	Image   Image  `json:"image"`
}

func ParseHexColor(hexStr string) (color color.RGBA) {
	color.A = 0xff
	if len(hexStr) == 7 {
		fmt.Sscanf(hexStr, "#%02x%02x%02x", &color.R, &color.G, &color.B)
	}
	if len(hexStr) == 4 {
		fmt.Sscanf(hexStr, "#%1x%1x%1x", &color.R, &color.G, &color.B)
	}

	if len(hexStr) == 9 {
		fmt.Sscanf(hexStr, "#%02x%02x%02x%02x", &color.R, &color.G, &color.B, &color.A)
	}

	if len(hexStr) == 5 {
		fmt.Sscanf(hexStr, "#%1x%1x%1x%1x", &color.R, &color.G, &color.B, &color.A)
	}

	return
}

func buildImageMessage(events []*Event, index int) *QWMessage {

	width := float64(600)
	height := 140 + float64(len(events)*140)

	for _, event := range events {
		if event.Action != Statistics {
			height += 30
		}
	}

	dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	gc := draw2dimg.NewGraphicContext(dest)
	gc.SetLineWidth(1)
	gc.SetFontData(draw2d.FontData{Name: "PingFang"})

	gc.Save()
	gc.SetFillColor(color.RGBA{255, 255, 255, 255})
	gc.ClearRect(0, 0, int(width), int(height))
	gc.Restore()

	offsetY := float64(0)

	// 写标题
	gc.Save()
	gc.SetFontSize(20)
	title := "腾讯云轻量服务器监控提醒通知"
	if index > 0 {
		title = fmt.Sprintf("%s(%d)", title, index)
	}
	left, _, right, _ := gc.GetStringBounds(title)

	gc.SetFillColor(color.Black)
	gc.FillStringAt(title, (width-(right-left))/2, 40)
	offsetY += 80
	gc.Restore()

	gc.Save()
	gc.SetFontSize(14)
	now := time.Now().Format("2006年01月02日 15:04")
	left, _, right, _ = gc.GetStringBounds(now)
	gc.SetFillColor(ParseHexColor("#5e6d82"))
	gc.FillStringAt(now, (width-(right-left))/2, offsetY)
	offsetY += 80
	gc.Restore()

	textColor := ParseHexColor("#5e6d82")
	progressBGColor := ParseHexColor("#ebeef5")

	for _, event := range events {

		// 主机名以及主机状态

		gc.Save()
		gc.SetFontSize(18)
		gc.SetFillColor(textColor)
		gc.FillStringAt(event.Name, 20, offsetY)
		left, _, right, _ := gc.GetStringBounds(event.Name)
		gc.Restore()

		gc.Save()
		gc.SetFontSize(10)
		cnState := GetCNState(event.State)
		gc.SetFillColor(ParseHexColor(GetStateColor(event.State)))
		offsetX := 35 + (right - left)
		left, _, right, _ = gc.GetStringBounds(cnState)
		draw2dkit.Rectangle(gc, offsetX, offsetY-18, offsetX+(right-left)+20, offsetY)
		gc.Fill()
		gc.SetFillColor(color.White)
		gc.FillStringAt(cnState, offsetX+10, offsetY-4)

		gc.Restore()

		offsetY += 30

		// 绘制流量使用情况

		text1 := fmt.Sprintf("流量一共%s，当前使用%s，%s", wellSize(event.Total), wellSize(event.Used),
			getTips(event.Used, event.Total))
		gc.Save()
		gc.SetFontSize(14)
		gc.SetFillColor(textColor)
		gc.FillStringAt(text1, 20, offsetY)
		offsetY += 20
		gc.Restore()

		// 绘制进度条
		gc.Save()
		gc.SetFillColor(progressBGColor)
		draw2dkit.Rectangle(gc, 20, offsetY, width-20, offsetY+20)
		gc.Fill()
		gc.Restore()

		gc.Save()
		gc.SetFillColor(ParseHexColor(getColor(event.Used, event.Total)))
		draw2dkit.Rectangle(gc, 20, offsetY, 20+(width-40)*float64(event.Used)/float64(event.Total), offsetY+20)
		gc.Fill()
		gc.Restore()

		offsetY += 50

		// 绘制操作
		if event.Action != Statistics {
			text2 := fmt.Sprintf("当前执行操作：【%s】", event.Action)
			gc.Save()
			gc.SetFontSize(14)
			gc.SetFillColor(textColor)
			gc.FillStringAt(text2, 20, offsetY)
			offsetY += 30
			gc.Restore()

		}

		offsetY += 40

	}

	buf := new(bytes.Buffer)
	jpeg.Encode(buf, dest, nil)

	imageData := buf.Bytes()
	base64Str := base64.StdEncoding.EncodeToString(imageData)
	ms5Str := fmt.Sprintf("%x", md5.Sum(imageData))

	message := QWMessage{
		MsgType: "image",
		Image: Image{
			Base64: base64Str,
			MD5:    ms5Str,
		},
	}

	return &message

}

func (n *QYWXNotifier) SendMessage(events []*Event) error {

	if len(events) == 0 {
		return nil
	}

	for i := 0; i < len(events); i += 5 {
		endIndex := i + 5
		if endIndex > len(events) {
			endIndex = len(events)
		}

		message := buildImageMessage(events[i:endIndex], i/5)
		sendData, _ := json.Marshal(&message)
		http.Post(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", n.key), "application/json",
			bytes.NewBuffer(sendData))
	}

	return nil
}
