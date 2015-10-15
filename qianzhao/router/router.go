package router

import (
	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/controller"
)

func Router(e *echo.Echo) {
	//// 秒速浏览器用户注册
	e.Post("/Api/v0.1/user/register", (&controller.User{}).Register)

	//// 秒速浏览器用户登录
	e.Post("/Api/v0.1/user/login", (&controller.User{}).Login)
	e.Get("/Api/v0.1/user/is_login", (&controller.User{}).IsLogin)
	e.Get("/Api/v0.1/user/logout", (&controller.User{}).LoginOut)

	//// 用户中心
	e.Get("/user", (&controller.User{}).MemberCenter)
	e.Get("/user/edit", (&controller.User{}).Edit)

	//// 用户收藏夹
	e.Post("/Api/v0.1/backup_favorite", (&controller.Favorite{}).BackupFavorite)
	e.Post("/Api/v0.1/get_favorite", (&controller.Favorite{}).GetFavorite)

	//// 宽带账号
	e.Post("/Api/v0.1/verify_Bandwith", (&controller.User{}).VerifyBandwith)
	e.Get("/Api/v0.1/get_Bandwith", (&controller.User{}).GetBandwith)

	//// 123
	e.Post("/Api/v0.1/speedup_open", (&controller.Ebit{}).SpeedupOpen)
	e.Get("/Api/v0.1/speedup_open", (&controller.Ebit{}).SpeedupOpen)

	e.Get("/Api/v0.1/speedup_open_check", (&controller.Ebit{}).SpeedupOpenCheck)
	e.Post("/Api/v0.1/speedup_open_check", (&controller.Ebit{}).SpeedupOpenCheck)

	e.Get("/Api/v0.1/speedup_check", (&controller.Ebit{}).SpeedupCheck)
	e.Post("/Api/v0.1/speedup_check", (&controller.Ebit{}).SpeedupCheck)

	//// 测速
	e.Get("/operate/speedup_prepare", (&controller.Operate{}).SpeedupPrepare)
	e.Get("/Api/v0.1/user/speed_test", (&controller.User{}).SpeedTest)
	e.Post("/operate/speedup_open_check", (&controller.Ebit{}).SpeedupOpenCheck)
	e.Post("/operate/speedup_open", (&controller.Ebit{}).SpeedupOpen)

	//// 接口
	e.Post("/interface/account", (&controller.Interfacer{}).AccountList)

	//// 宽带（双速网）
	e.Post("/broadbrand/start", (&controller.BroadBand{}).Start)
	e.Get("/broadbrand/resettime", (&controller.BroadBand{}).ResetTime)
	e.Get("/version/update", (&controller.Index{}).Update)

	//// 统计
	e.Get("/download", (&controller.Statistics{}).Download)
	e.Post("/stats/day", (&controller.Statistics{}).DayActivity)
	e.Post("/stats/sidbar", (&controller.Statistics{}).SideBar)

	//// 反馈
	e.Get("/feedback", (&controller.FeedBack{}).Index)
	e.Post("/feedback/post", (&controller.FeedBack{}).Post)
	e.Get("/feedback/pic", (&controller.FeedBack{}).Pic)
}

//// 秒速浏览器用户注册
//Route::post("Api/v0.1/user/register", "UserController@register");

//// 秒速浏览器用户登录
//Route::post("Api/v0.1/user/login", "UserController@login");
//Route::get("Api/v0.1/user/is_login", "UserController@is_login");
//Route::get("Api/v0.1/user/logout", function(){
//    Auth::logout();
//    $return['code'] = "200";
//    $return['msg'] = "用户已注销!";
//    return Response::json($return, 200, [], JSON_UNESCAPED_UNICODE)->header('Content-Type', "application/json;charset=UTF-8");
//});

//// 用户中心
////Route::get("MemberCenter",  array('before' => 'is_login', 'uses' => 'UserController@member_center'));
//Route::get('user', 'UserController@member_center');
//Route::get('user/edit', 'UserController@edit');

//// 用户收藏夹
//Route::post("Api/v0.1/backup_favorite", array('before' => 'is_login', 'uses' => 'FavoriteController@backup_favorite'));
//Route::post("Api/v0.1/get_favorite",  array('before' => 'is_login', 'uses' => 'FavoriteController@get_favorite'));

////
//Route::post('Api/v0.1/verify_Bandwith', 'UserController@verify_Bandwith');
//Route::get('Api/v0.1/get_Bandwith', 'UserController@get_Bandwith');

//// 123
//Route::any('Api/v0.1/speedup_open_check', 'EbitController@speedup_open_check');
//Route::any('Api/v0.1/speedup_check', 'EbitController@speedup_check');
//Route::any('Api/v0.1/speedup_open', 'EbitController@speedup_open');

//// 测速
//Route::get('Api/v0.1/user/speed_test', 'UserController@speed_test');

//Route::get("/operate/speedup_prepare", "OperateController@speedup_prepare");
//Route::post("/operate/speedup_open_check", "EbitController@speedup_open_check");
//Route::post("/operate/speedup_open", "EbitController@speedup_open");
