package controller

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao/common/captcha"
	"github.com/qgweb/gopro/qianzhao/common/function"
	"github.com/qgweb/gopro/qianzhao/common/session"
	"github.com/qgweb/gopro/qianzhao/model"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"
)

type FeedBack struct {
}

func (this *FeedBack) Index(ctx *echo.Context) error {
	fb := model.FeedBack{}
	ref := reflect.TypeOf(fb)
	num := ref.NumField()

	for i := 0; i < num; i++ {
		log.Error(ref.Field(i).Tag.Get("json"))
	}

	return ctx.Render(200, "feedback_index", "")
}

func (this *FeedBack) Post(ctx *echo.Context) error {
	var (
		fbmodel = &model.FeedBack{}
		err     error
	)

	// session
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}
	defer sess.SessionRelease(ctx.Response())

	// bind
	fbmodel.Btype = convert.ToInt(ctx.Form("type"))
	fbmodel.QDescribe = ctx.Form("qdescribe")
	fbmodel.Qtype = convert.ToInt(ctx.Form("qtype"))
	fbmodel.Contact = ctx.Form("contact")
	fbmodel.Tcontact = convert.ToInt(ctx.Form("tcontact"))

	// check data
	err = fbmodel.CheckData(fbmodel)
	if err != nil {
		return ctx.HTML(200, Alert(err.Error()))
	}

	if ctx.Form("checkcode") == "" {
		return ctx.HTML(200, Alert("验证码不能为空"))
	}

	if sess.Get("code") != nil && ctx.Form("checkcode") != sess.Get("code").(string) {
		return ctx.HTML(200, Alert("验证码错误"))
	}

	fbmodel.Qpic, err = UploadPic(ctx)
	if err != nil {
		log.Error(err)
	}

	if fbmodel.AddRecord(fbmodel) {
		return ctx.HTML(200, Alert("反馈成功!"))
	} else {
		return ctx.HTML(200, Alert("反馈失败!"))
	}
}

func (this *FeedBack) Pic(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}

	defer sess.SessionRelease(ctx.Response())

	code := captcha.Get(ctx.Response(), ctx.Request())
	sess.Set("code", code)
	return nil
}

//  upload pic
func UploadPic(ctx *echo.Context) (string, error) {
	pf, ph, err := ctx.Request().FormFile("qpic")
	if err != nil {
		log.Error(err)
		return "", err
	}

	ext := filepath.Ext(ph.Filename)
	fileName := fmt.Sprintf("%s/public/upload/feedback/%s%s", function.GetBasePath(), time.Now().Format("20060102130405"), ext)
	f, err := os.Create(fileName)
	if err != nil {
		log.Error(err)
		return "", nil
	}

	io.Copy(f, pf)
	return strings.Replace(fileName, function.GetBasePath(), "", -1), nil
}

func Alert(err string) string {
	buf := bytes.NewBuffer([]byte{})
	t, _ := template.New("xxx").Parse(`
		<script type="text/javascript">
		alert("{{.}}")
		window.location="/feedback";
		</script>
		`)
	t.Execute(buf, err)
	return buf.String()
}
