$(function () {
    $('.rememberPwd_img,.rememberPwd_text').click(function () {//记住密码
        var rmbimg = $('.rememberPwd_img');
        rmbimg.hasClass('checked') ? rmbimg.removeClass('checked') : rmbimg.addClass('checked');
        $('#rememberPwd')[0].checked = rmbimg.hasClass('checked');
    });
    $('a.go_reg').click(function () {//转到注册页
        if (parent != window) {
            parent.$('body').trigger('topage', 'register');
            return false;
        }
    })
    $('a.go_login').click(function () {//转到注册页
        if (parent != window) {
            parent.$('body').trigger('topage', 'login');
            return false;
        }
    })
    $('body').on('topage', function (e, p) {
        if (p == 'register') {
            $('#register').addClass('show').removeClass('hide');
            $('#login').addClass('hide').removeClass('show');
        } else if (p == 'login') {
            $('#login').addClass('show').removeClass('hide');
            $('#register').addClass('hide').removeClass('show');
        }
        doAnmi();
    })
    $('.btn_login').click(function () {//点击登录
        var username = $('#username').val();
        var password = $('#password').val();
        var isRmb = $('#rememberPwd')[0].checked ? 1 : 0;

        var err = [];
        !(username && password) && err.push('用户名,密码不能为空');
        err.length ? errshow(err.join(';')) : $.ajax({
            url: '/user/login',
            data: {username: username, password: password, isRmb: isRmb},
            type: 'post',
            dataType: 'json',
            success: function (d) {
                if (d.code == '200') {//注册成功
//                    location.href = '/user/login'
                    errshow(d.msg);
                } else {
                    errshow(d.msg);
                }
            }
        })
        return false;
    })

    $('.btn_reg').click(function () {//点击注册
        var username = $('#username').val();
        var password = $('#password').val();
        var pwdAgain = $('#pwdAgain').val();
        var err = [];
        !(username && password) && err.push('用户名,密码不能为空');
        password != pwdAgain && err.push('两次输入的密码不一致');
        err.length ? errshow(err.join(';')) : $.ajax({
            url: '/user/register',
            data: {username: username, password: password, pwd: pwdAgain},
            type: 'post',
            dataType: 'json',
            success: function (d) {
                if (d.code == '200') {//注册成功
//                    location.href = '/user/login';
                    parent.$('body').trigger('topage', 'login');
                    errshow(d.msg);
                } else {
                    errshow(d.msg);
                }
            }
        })
        return false;
    })
})
function errshow(str) {//显示错误信息
    $('#errinfo').text(str);
}
function doAnmi() {//执行旋转动画
    var s = $('#pages iframe.show').css('WebkitTransform', 'rotateY(270deg)');
    var h = $('#pages iframe.hide');

    function setH(i, cbk) {
        h.css('WebkitTransform', 'rotateY(' + i + 'deg)');
        setTimeout(function () {
            i++;
            i <= 90 ? setH(i, cbk) : cbk();
        }, 2);
    }

    function setS(i, cbk) {
        s.css('WebkitTransform', 'rotateY(' + i + 'deg)');
        setTimeout(function () {
            i++;
            i <= 360 ? setS(i, cbk) : cbk();
        }, 2);
    }

    setH(0, function () {
        setS(270, function () {
            s.css('WebkitTransform', 'rotateY(0deg)');
        });
    });
}