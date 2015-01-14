package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/models"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

type UsersWebController struct {
	beego.Controller
}

func (this *UsersWebController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *UsersWebController) PostGravatar() {
	//从请求中读取图片信息，图片保存在相应
	file, fileHeader, err := this.Ctx.Request.FormFile("file")
	if err != nil {
		beego.Error(fmt.Sprintf("[image] 处理上传头像错误,err=%s", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"图片上传处理失败\"}"))
		return
	}

	//读取文件后缀名，如果不是图片，则返回错误
	prefix := strings.Split(fileHeader.Filename, ".")[0]
	suffix := strings.Split(fileHeader.Filename, ".")[1]
	if suffix != "png" && suffix != "jpg" && suffix != "jpeg" {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"文件的扩展名必须是jpg、jpeg或者png!\"}"))
		return
	}

	f, err := os.OpenFile(fmt.Sprintf("%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", fileHeader.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		//处理文件错误
		beego.Error(fmt.Sprintf("[image] 处理上传头像错误,err=%s", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"图片上传处理失败\"}"))
		return
	}
	io.Copy(f, file)
	f.Close()

	// decode jpeg into image.Image
	var img image.Image
	imageFile, err := os.Open(fmt.Sprintf("%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", fileHeader.Filename))
	if err != nil {
		beego.Error(fmt.Sprintf("[image] 上传图片预失败,err=%s", err))
	}
	switch suffix {
	case "png":
		img, err = png.Decode(imageFile)
	case "jpg":
		img, err = jpeg.Decode(imageFile)
	case "jpeg":
		img, err = jpeg.Decode(imageFile)
	}
	if err != nil {
		beego.Error(fmt.Sprintf("[image] 裁剪图片失败,err=%s", err))
		imageFile.Close()
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"图片上传处理失败\"}"))
		return
	}
	imageFile.Close()
	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(100, 100, img, resize.Lanczos3)

	out, err := os.Create(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix))
	if err != nil {
		beego.Error(fmt.Sprintf("[image] 裁剪图片失败,err=%s", err))
	}
	defer out.Close()
	// write new image to file
	switch suffix {
	case "png":
		png.Encode(out, m)
	case "jpg":
		jpeg.Encode(out, m, nil)
	case "jpeg":
		jpeg.Encode(out, m, nil)
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"文件上传成功！\",\"url\":\"" + fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix) + "\"}"))
	return
}

func (this *UsersWebController) GetProfile() {
	//加载session
	user, ok := this.GetSession("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"session加载失败\",\"url\":\"/auth\"}"))
		return
	}
	user2json, err := json.Marshal(user)
	if err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] session解码json失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"session解码json失败\",\"url\":\"/auth\"}"))
		return
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(user2json)
	return
}