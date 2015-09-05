package controller

import (
	"log"
	//"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/common/session"
	//"golang.org/x/crypto/bcrypt"
)

func Index(ctx *echo.Context) error {
	//$2y$10$PJwVP4S3B9QBcai/bP.NA.ujSX8ue90wDbPm1B423wtOeiBVyWpFG
	//$2a$10$dBBYYkpONL1rnBQ2BH2Hy.emfcgOKv4NZCP9o1SwIBT6d1spUMPNq
	return ctx.Render(200, "usercenter", "333")
}

func Show(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Println("获取session失败：", err)
	}

	defer sess.SessionRelease(ctx.Response())

	//$2y$10$PJwVP4S3B9QBcai/bP.NA.ujSX8ue90wDbPm1B423wtOeiBVyWpFG
	//$2a$10$dBBYYkpONL1rnBQ2BH2Hy.emfcgOKv4NZCP9o1SwIBT6d1spUMPNq
	return ctx.String(http.StatusOK, sess.Get("NAME").(string))
}
