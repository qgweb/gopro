package session

import (
	"github.com/astaxie/beego/session"
	"github.com/labstack/echo"
)

var globalSessions *session.Manager

func init() {
	globalSessions, _ = session.NewManager("memory", `{"cookieName":"211susessionid", "enableSetCookie,omitempty": true, "gclifetime":3600, "maxLifetime": 86400, "secure": false, "sessionIDHashFunc": "sha1", "sessionIDHashKey": "", "cookieLifeTime": 86400, "providerConfig": ""}`)
	go globalSessions.GC()
}

func GetSession(ctx *echo.Context) (session.SessionStore, error) {
	sess, err := globalSessions.SessionStart(ctx.Response(), ctx.Request())
	if err != nil {
		return nil, err
	}
	return sess, err
}
