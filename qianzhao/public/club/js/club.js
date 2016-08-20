/*
 *  会员俱乐部js
 */

var lastAngel = null;
laytpl.config({open: '{#{', close: '}#}'});

function rotateFunc(obj, angle, fn) {//
    var rouNum = Math.ceil(Math.random() * 5.5) + 4.5;
    obj.rotate({
        angle: lastAngel || 0, //从什么角度开始转
        duration: 1000 * rouNum, //转多久
        animateTo: angle + 720 * rouNum, //转几度(一圈360度)
        callback: function () {//转完后执行回调
            lastAngel = angle;
            // $('#clickRate').show();
            fn && fn();
        }
    })
}

$(function () {
    $('body').click(function () {
        $('.popup').hide();
    })
    $('body').on('click', 'a', function () {
        if ($(this).attr('href') == '#') {
            return false;
        }
    })
    //滚动
    function scrollT(ele) {
        var ulH = $('.' + ele).find('ul').height();
        if (ulH > 150) {
            var htmlT = $('.' + ele).html();
            $('.' + ele).append(htmlT);
            var li = 0;
            setInterval(function () {
                if (li >= ulH / 30) {
                    li = 0;
                    $('.' + ele).find('ul').eq(0).remove();
                    $('.' + ele).append(htmlT)
                }
                $('.' + ele).find('ul').eq(0).css({
                    'marginTop': -30 * li + 'px'
                });
                li++;
            }, 800);
        }
    }

    scrollT('record_lottery');
    scrollT('record_phone');

    //弹出框关闭
    $('.popup input[type=button]').click(function () {
        $('.popup').hide();
    });

})

//签到
$(function () {
    function sign(his) {
        var sign = his.split("");
        if (sign.length > 0) {
            for (var i = 0; i < sign.length; i++) {
                var liobj = $('.container02 ul li');
                if (sign[i] == "1") {
                    liobj.eq(i).addClass("sign").html("<i></i><span>已签到</span>");
                } else {
                    liobj.eq(i).removeClass("sign").html("第" + (i + 1) + "天");
                }
            }
        }
        if (sign.length == 5) {
            $('.sign_btn').addClass('already');
        }
    }

    sign($('#signhistory').val());
    $('.sign_btn').click(function (e) {
        if ($(e).attr("class").indexof("already") != -1) {
            return;
        }
        $(e).attr("css").contains("already")
        $.ajax({
            'url': "/club/sign",
            'dataType': 'json',
            'success': function (data) {
                if (data.code == "301" || data.ret == "-1") {
                    layer.msg(data.msg, {icon: 5});
                    return;
                }
                if (data.ret == "0" && data.data) {
                    sign(data.data);
                    $('.sign_btn').addClass('already');
                    if (data.data == "11111") {
                        $('.popup').hide();
                        $('#pprompt').show();
                        $('#pprompt').children().show();
                        $('.lottery b').html($('.lottery b').text() * 1 + 1);
                        return;
                    }

                }
            }
        })
    })
})

// 猜字谜
$(function () {
    $('.guess_btn').click(function () {
        var an = $('.guess_an').val();
        if (!an) {
            layer.msg("谜底不能为空!", {icon: 5});
            return;
        }
        $.ajax({
            'url': "/club/word",
            'dataType': 'json',
            'type': "post",
            'data': {w: an},
            'success': function (data) {
                if (data.code == "301" || data.ret == "-1") {
                    layer.msg(data.msg, {icon: 5});
                    return;
                }
                if (data.ret == "0" && data.data) {
                    var msgData = {
                        "0": " 本轮话费已抢完，请明日继续",
                        "1": "恭喜您抽中5元话费充值卡！",
                        "2": "恭喜您抽中10元话费充值卡！",
                        "3": "恭喜您抽中20元话费充值卡！",
                    }
                    if (msgData[data.data.n]) {
                        $('.popup').hide();
                        $('.popup .wtips_list p').html(msgData[data.data.n]);
                        $('.popup .wtips_list dd a').html(data.data.c);
                        $('#pwin').show();
                    }
                    return;
                }
            }
        })
    })
})

//转盘
$(function () {
    $('.draw_chassis div').on('click', function () {
        var ran = Math.random();
        $.ajax({
            url: '/club/tun',
            type: 'get',
            dataType: 'json',
            success: function (d) {
                var bc = $('.lottery b').text() * 1 - 1;
                $('.lottery b').html(bc < 0 ? 0 : bc);
                if (d.code == 200) {
                    var resl;
                    var html;
                    switch (d.num) {
                        case 0:
                            if (ran > 0.5) {
                                resl = 2;
                            } else {
                                resl = 6;
                            }
                            html = "谢谢参与";
                            break;
                        case 1:
                            if (ran > 0.5) {
                                resl = 1;
                            } else {
                                resl = 5;
                            }
                            html = "1元话费充值卡";
                            break;
                        case 2:
                            if (ran > 0.5) {
                                resl = 0;
                            } else {
                                resl = 4;
                            }
                            html = '5元话费充值卡';
                            break;
                        case 3:
                            if (ran > 0.5) {
                                resl = 3;
                            } else {
                                resl = 7;
                            }
                            html = '10元话费充值卡';
                            break;
                    }
                    rotateFunc($('.draw_chassis div'), 45 * resl + 22.5, function () {
                        $('.popup').hide();
                        $('.popup .wtips_list p').html(d.res);
                        $('.popup .wtips_list dd a').html(d.rcode);
                        $('#pwin').show();
                        return;
                    })
                } else if (d.code == 301 || d.ret == -1) {
                    layer.msg(d.msg, {icon: 5});
                }
            }
        })
    })
})

//我的记录
$(function () {
    $(".precord_c").click(function () {
        return false;
    })
    $(".view_record").click(function () {
        $.ajax({
            'url': "/club/mrecord",
            'dataType': 'json',
            'type': 'post',
            'data': {p: 1},
            'success': function (data) {
                if (data.code == "301") {
                    layer.msg(data.msg, {icon: 5});
                    return;
                }
                $('.popup').hide();
                laytpl($('#mlist').html()).render(data.slice(0, 3), function (html) {
                    $('.precord_list table tbody').html(html);
                });
                //分页栏
                var h = '<a href="#" tg="1">上一页</a>';
                for (var i = 0; i < Math.ceil(data.length / 3); i++) {
                    var s = 'class="current"';
                    h += '<a href="#" tg="o" ' + (i == 0 ? s : "") + '>' + (i + 1) + '</a>';
                }
                h += '<a href="#" tg="2">下一页</a>';
                $('.page_list').html(h);
                $('#precord').show();
                window.my_record = data;

                $('.page_list a').click(function () {
                    var obj = $('.page_list a');
                    var cindex = 0;
                    var size = 3;
                    var total = Math.ceil(window.my_record.length / 3);
                    var pobj = null;
                    var nobj = null;
                    var sobj = this;

                    obj.each(function (index, dom) {
                        if ($(dom).text() == "上一页") {
                            pobj = function (obj) {
                                return $(obj);
                            }(dom);
                        }
                        if ($(dom).text() == "下一页") {
                            nobj = function (obj) {
                                return $(obj);
                            }(dom);
                        }
                    })

                    obj.removeClass("current");
                    if ($(this).attr('tg') == "o") {
                        $(this).addClass("current");
                        cindex = $(this).text();
                    }
                    if ($(this).attr("tg") != "o") {
                        cindex = $(this).attr("tg");
                        obj.each(function (index, dom) {
                            if ($(dom).text() == cindex) {
                                $(dom).addClass("current");
                                return;
                            }
                        })
                    }
                    cindex = cindex * 1;
                    var pcindex = cindex - 1 <= 0 ? 1 : cindex - 1;
                    var ncindex = cindex + 1 >= total ? total : cindex + 1;
                    pobj.attr("tg", pcindex);
                    nobj.attr("tg", ncindex);

                    if (window.my_record) {
                        var sdata = window.my_record.slice((cindex - 1) * size, cindex * size);
                        laytpl($('#mlist').html()).render(sdata, function (html) {
                            $('.precord_list table tbody').html(html);
                        });
                    }
                })
            }
        })
    })
})