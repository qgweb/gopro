/**
 * Created by zb on 16-8-24.
 */
//切换2套
$(function () {
    $(window).resize(function () {
        if (document.body.clientWidth < 1368) {
            $('#link').attr("href", "./main1/css/style.css");
            $('#bimg1').attr("src", "./main1/img/banner01.jpg");
            $('#bimg2').attr("src", "./main1/img/banner02.jpg");
        } else {
            $('#link').attr("href", "./main/css/style.css");
            $('#bimg1').attr("src", "./main/img/banner01.jpg");
            $('#bimg2').attr("src", "./main/img/banner02.jpg");
        }

    })
    if (document.body.clientWidth < 1368) {
        $('#link').attr("href", "./main1/css/style.css");
        $('#bimg1').attr("src", "./main1/img/banner01.jpg");
        $('#bimg2').attr("src", "./main1/img/banner02.jpg");
    }
})

$(function () {
    //时间
    var h = new Date();
    var wstr = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"];
    $('.date').html(h.getMonth() + "月" + h.getDate() + "日&nbsp;" + wstr[h.getDay()]);
    $('body').on('click', 'a', function () {
        if ($(this).attr('href') == '#') {
            return false;
        }
    }).on('focus', 'a', function () {
        $(this).blur();
    });
    if (document.body.clientHeight > 595) {

    }
    $('.website_tab li').eq(4).css('border-right-width', 0);
    $('.module ul').each(function () {
        $('.module ul').eq(1).css('border-top', '1px solid #eee');
        $(this).find('li').eq(3).css('border-right-width', 0);
    });

    $('.website_tab li a').click(function () {
        $('.website_tab li a').removeClass('current');
        $(this).addClass('current');
        $('.website_list').hide().eq($(this).parent().index()).show();
    });
})
//收藏
function AddFavorite() {
    try {
        window.external.addFavorite(location.href, "千兆浏览器首页");
    }
    catch (e) {
        try {
            window.sidebar.addPanel(title, url, "");
        }
        catch (e) {
            alert("抱歉，您所使用的浏览器无法完成此操作。\n\n加入收藏失败，请使用Ctrl+D进行添加");
        }
    }
}
//banner轮播
function banner() {
    var timer = null;
    var inx = 0;
    var len = $('.banner_list').find('li').length;

    function next() {
        inx++;
        tab();
    }

    function tab() {
        if (inx >= len) {
            inx = 0;
        }
        if (inx == -1) {
            inx = len - 1;
        }
        var w = $('.banner_list').find('li').width();
        startMove($('.banner_list')[0], {'marginLeft': -w * inx});
        $('.banner_list').find('li').removeClass('current').eq(inx).addClass('current');
        $('.banner_point').find('a').removeClass('active').eq(inx).addClass('active');
    }

    timer = setInterval(next, 5000);
    $('.banner_point a').mouseover(function () {
        clearInterval(timer);
        inx = $(this).index();
        tab();
    })
}
if ($('.banner_list li').length) {
    banner();
}
var startMove = function (obj, json, options) {
    options = options || {};
    options.time = options.time || 700;
    options.type = options.type || 'ease-out';
    if (obj != undefined) {
        clearInterval(obj.timer);
        var count = Math.floor(options.time / 30);
        var start = {};
        var dis = {};
        for (var name in json) {
            if (name == 'opacity') {
                start[name] = Math.round(parseFloat($(obj).css(name)) * 100);
            } else {
                start[name] = parseInt($(obj).css(name));
            }
            dis[name] = json[name] - start[name];
        }
        var n = 0;
        obj.timer = setInterval(function () {
            n++;
            for (var name in json) {
                switch (options.type) {
                    case 'linear':
                        var a = n / count;
                        var cur = start[name] + dis[name] * a;
                        break;
                    case 'ease-in':
                        var a = n / count;
                        var cur = start[name] + dis[name] * a * a * a;
                        break;
                    case 'ease-out':
                        var a = 1 - n / count;
                        var cur = start[name] + dis[name] * (1 - a * a * a);
                        break;
                }
                if (name == 'opacity') {
                    obj.style.opacity = cur / 100;
                    obj.style.filter = 'alpha(opacity:' + cur + ')';
                } else {
                    obj.style[name] = cur + 'px';
                }
            }
            if (n == count) {
                clearInterval(obj.timer);
                options.end && options.end();
            }
        }, 50);
    }
    // clearInterval(obj.timer);
};

// 活动banner
$(function () {
    if ($.cookie("bannerhide") != "1") {
        $('.haibao').animate({
            width: '100%',
            height: '100%',
            left: 0,
            top: 0
        }, 1500)
        $('.haibao img').animate({
            width: '1024px',
            height: '600px',
            marginLeft: '-512px',
            marginTop: '-300px'
        }, 1500)
        $('.close').click(function () {
            $('.haibao').hide();
            $.cookie("bannerhide", "1");
        });
        setTimeout(function () {
            $('.close').fadeIn('slow');
        }, 1500)
        setTimeout(function () {
            $('.haibao').fadeOut(1000);
        }, 5000)
    }
})

// 历史记录
function parseURL(url) {
    var a = document.createElement('a');
    a.href = url;
    return {
        source: url,
        protocol: a.protocol.replace(':', ''),
        host: a.hostname,
        port: a.port,
        query: a.search,
        params: (function () {
            var ret = {},
                seg = a.search.replace(/^\?/, '').split('&'),
                len = seg.length, i = 0, s;
            for (; i < len; i++) {
                if (!seg[i]) {
                    continue;
                }
                s = seg[i].split('=');
                ret[s[0]] = s[1];
            }
            return ret;
        })(),
        file: (a.pathname.match(/\/([^\/?#]+)$/i) || [, ''])[1],
        hash: a.hash.replace('#', ''),
        path: a.pathname.replace(/^([^\/])/, '/$1'),
        relative: (a.href.match(/tps?:\/\/[^\/]+(.+)/) || [, ''])[1],
        segments: a.pathname.replace(/^\//, '').split('/')
    };
}
function GetHistoryList() {
    console.log("GetHistoryList");
    chrome.send("GetHistoryList");
}

function GetHistoryListDone(entries) {
    var ii = 0;
    for (var i = 0; i < entries.length; i++) {
        var et = entries[i];
        console.log(et)
        if (et.title == "" || et.title.indexOf("千兆") != -1) {
            continue;
        }

        ii++
        if (ii > 6) {
            return;
        }
        var obj = parseURL(et.url);
        var html = '<li><a href="' + obj.source +
            '" target="_blank" title=""><img onerror="this.src=\'http://qianzhao.221su.com/main/img/ico.png\'" src="' + (obj.protocol + "://" + obj.host + "/favicon.ico") +
            '" alt=""><span>' + et.title + '</span></a></li>'
        $('.often_list').append(html)
    }
}

$(function () {
    GetHistoryList()
})
