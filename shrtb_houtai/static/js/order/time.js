function model(obj, act, top, stats) {
    if (act == 'show') {
        stats = stats || 'fixed',
            top = top || 300;
        if (top == 1) {
            top = 0
        }
        var h = $(obj).height();
        var w = $(obj).width();
        $(obj).css({
            'position': stats,
            'left': '50%',
            'marginLeft': -w / 2 + 'px',
            'top': -h + 'px',
            'zIndex': '10000001'
        }).show();
        startMove($(obj)[0], {
            top: top
        })
        if (!$('.time_shade').length) {
            $(obj).parent().append('<div class="time_shade"></div>');
        } else {
            $('.time_shade').show();
        }
    } else if (act == 'hide') {
        var t;
        if ($(obj).offset()) {
            t = $(obj).offset().top;
        }
        // var t=$(obj).offset().top;
        t = top || 0;
        // $(obj).animate({
        //     top:t+'px',
        //     opacity:0
        // },function(){
        //     $('.time_shade').hide();
        //     $(obj).hide();
        //     $(obj).css({
        //         opacity:1
        //     })
        // })
        startMove($(obj)[0], {
            top: t,
            opacity: 0
        }, {
            end: function() {
                $('.time_shade').hide();
                $(obj).hide();
                $(obj).css({
                    opacity: 1
                })
            }
        })
    }
}
var startMove = function(obj, json, options) {
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
        obj.timer = setInterval(function() {
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
        }, 30);
    }
    // clearInterval(obj.timer);
};
window.allHourCache = $('#rbwhs').find('div.week div.hour');

function setMyPrdTime(tStr, ub) { //设置显示的投放时段---tStr的格式"11111111111111111|11111111111111|010101001|010101001|222222|11111111|00000000"
    console.log(tStr)
    var WeekAry = tStr.split('|');
    var stuAry = ['stop', 'allow', 'unable'];
    var wh = allHourCache.removeClass('allow').removeClass('stop');
    if (!ub) {
        wh.removeClass('unable');
    }
    for (var x in WeekAry) {
        WeekAry[x] = WeekAry[x].split('');
    }
    for (var i = 0; i < 7; i++) {
        for (var j = 0; j < 24; j++) {
            $('#week' + i + 'hour' + j).not('.unable').addClass(stuAry[WeekAry[i][j]]);
        }
    }
}

$('#setPrdTime').click(function() { //点击自定义按钮,显示时段选择界面
    setMyPrdTime($('#time_point').val());
    $('#periodTime').modal('show');
    model('#periodTime', 'show');
})

function sumSelect() {
    var stms = $('#time_point').val().replace(/\|/g, '').split('');
    var sumOne = 0;
    for (var i in stms) {
        stms[i] == '1' && (sumOne++);
    }
    var persent = ((sumOne / (stms.length)) * 100).toFixed(1);
    // console.log($('#allowPercent')[0])
    $('#allowPercent').html(persent + '%');
}
$(function() {
    sumSelect();
    $('#surwMyPrds').click(function() {
        $('#time_point').val(getPagePrdDatas());
        $('#allowPercent').text(((($('#rbwhs .hour.allow').length) / ($('#rbwhs .hour').length)) * 100).toFixed(2));
        model('#periodTime', 'hide');
    })
})