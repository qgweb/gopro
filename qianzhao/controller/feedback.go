package controller

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/nfnt/resize"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao/common/captcha"
	"github.com/qgweb/gopro/qianzhao/common/function"
	"github.com/qgweb/gopro/qianzhao/common/session"
	"github.com/qgweb/gopro/qianzhao/model"
	"image/jpeg"
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

func (this *FeedBack) Index(ctx echo.Context) error {
	fb := model.FeedBack{}
	ref := reflect.TypeOf(fb)
	num := ref.NumField()

	for i := 0; i < num; i++ {
		log.Error(ref.Field(i).Tag.Get("json"))
	}

	return ctx.Render(200, "feedback_index", "")
}

func (this *FeedBack) Post(ctx echo.Context) error {
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
	defer sess.SessionRelease(ctx.Response().(*standard.Response).ResponseWriter)

	// bind
	fbmodel.Btype = convert.ToInt(ctx.FormValue("type"))
	fbmodel.QDescribe = ctx.FormValue("qdescribe")
	fbmodel.Qtype = convert.ToInt(ctx.FormValue("qtype"))
	fbmodel.Contact = ctx.FormValue("contact")
	fbmodel.Tcontact = convert.ToInt(ctx.FormValue("tcontact"))

	// check data
	err = fbmodel.CheckData(fbmodel)
	if err != nil {
		return ctx.HTML(200, Alert(err.Error()))
	}

	if ctx.FormValue("checkcode") == "" {
		return ctx.HTML(200, Alert("验证码不能为空"))
	}

	if sess.Get("code") != nil && ctx.FormValue("checkcode") != sess.Get("code").(string) {
		return ctx.HTML(200, Alert("验证码错误"))
	}

	fbmodel.Qpic, err = UploadPic(ctx, "feedback", false)
	if err != nil {
		log.Error(err)
	}

	if fbmodel.AddRecord(fbmodel) {
		return ctx.HTML(200, Alert("反馈成功!"))
	} else {
		return ctx.HTML(200, Alert("反馈失败!"))
	}
}

func (this *FeedBack) Pic(ctx echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}

	defer sess.SessionRelease(ctx.Response().(*standard.Response).ResponseWriter)

	code := captcha.Get(ctx.Response().(*standard.Response).ResponseWriter,
		ctx.Request().(*standard.Request).Request)
	sess.Set("code", code)
	return nil
}

//  upload pic
func UploadPic(ctx echo.Context, path string, isthum bool) (string, error) {
	fh, err := ctx.FormFile("qpic")
	if err != nil {
		log.Error(err)
		return "", errors.New("图片上传失败")
	}

	ext := filepath.Ext(fh.Filename)
	if !(ext == ".png" || ext == ".jpg") {
		return "", errors.New("图片格式不正确")
	}
	fileName := fmt.Sprintf("%s/public/upload/%s/%s%s", function.GetBasePath(), path,
		time.Now().Format("20060102130405"), ext)
	f, err := os.Create(fileName)
	if err != nil {
		log.Error(err)
		return "", nil
	}
	defer f.Close()

	nf, err := fh.Open()
	if err != nil {
		log.Error(err)
		return "", nil
	}
	if isthum {
		img, err := jpeg.Decode(nf)
		if err == nil {
			m := resize.Resize(150, 150, img, resize.Lanczos3)
			jpeg.Encode(f, m, nil)
		}
	} else {
		io.Copy(f, nf)
	}

	return strings.Replace(fileName, function.GetBasePath()+"/public", "", -1), nil
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
