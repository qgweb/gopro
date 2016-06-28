window.allHourCache = $('#rbwhs').find('div.week div.hour'); //缓存所有的时间点,不用每次都用jquery获取一遍   提高性能
/*
 * 0---暂停    1---投放    2---不可选
 * */
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

function changeOnePrd(prd) {
    if (!prd.hasClass('unable')) {
        if (prd.hasClass('stop')) {
            prd.removeClass('stop').addClass('allow');
        } else if (prd.hasClass('allow')) {
            prd.removeClass('allow').addClass('stop');
        }
        $('#nowSet').filter(':not(:checked)').each(function() {
            this.click();
        })
    }
}

function getPagePrdDatas() {
    var data = [];
    $('#rbwhs .week').each(function() {
        var singleWeek = '';
        $(this).find('.hour').each(function(i, v) {
            var $t = $(v);
            var ts = 0;
            if ($t.hasClass('allow')) {
                ts = 1;
            } else if ($t.hasClass('stop')) {
                ts = 0;
            } else if ($t.hasClass('unable')) {
                ts = 2;
            }
            singleWeek += ts;
        })
        data.push(singleWeek);
    })
    return data.join('|');
}
$(function() {
        //判断是否只读
        if (!window.isOnlyShow) {
            $('#allDays').click(function() { //整周全天
                setMyPrdTime('111111111111111111111111|111111111111111111111111|111111111111111111111111|111111111111111111111111|111111111111111111111111|111111111111111111111111|111111111111111111111111', true);
            });
            $('#workDays').click(function() { //工作日
                setMyPrdTime('111111111111111111111111|111111111111111111111111|111111111111111111111111|111111111111111111111111|111111111111111111111111|000000000000000000000000|000000000000000000000000', true);
            })
            $('#restDays').click(function() { //休息日
                setMyPrdTime('000000000000000000000000|000000000000000000000000|000000000000000000000000|000000000000000000000000|000000000000000000000000|111111111111111111111111|111111111111111111111111', true);
            })
            $('#rbwhs').find('div.week div.hour').click(function() { //点击某一个时间点
                    changeOnePrd($(this));
                })
                /*
                 * 拖动批量修改
                 * */
            allHourCache.mousedown(function() {
                allHourCache.filter('.startHour').removeClass('startHour');
                $(this).addClass('startHour');
                window.startChange = true;
            }).mouseover(function() {
                if (window.startChange) {
                    allHourCache.filter('.endHour').removeClass('endHour');
                    $(this).addClass('endHour');
                    manyChange();
                }
            })
            $('body').mouseup(function() {
                allHourCache.filter('.ls').each(function() {
                    changeOnePrd($(this).removeClass('ls'));
                })
                window.startChange = false;
            })
        } else {
            $('#pbdCont .cue').html('&nbsp;');
            $('#surwMyPrds').remove();
            $('#nowSet,#allDays,#workDays,#restDays').each(function() {
                this.readOnly = true;
                this.disabled = true;
            })
        }
    })
    // $('#surwMyPrds').click(function() {
    //     //$('#time_point').val(getPagePrdDatas());
    //     // model('#periodTime', 'hide');
    // })

function manyChange() {
    var idStart = allHourCache.filter('.startHour').attr('id');
    var idEnd = allHourCache.filter('.endHour').attr('id');
    var sWeek = idStart.split('hour')[0].replace('week', '');
    var sHour = idStart.split('hour')[1];
    var eWeek = idEnd.split('hour')[0].replace('week', '');
    var eHour = idEnd.split('hour')[1];
    var weekAry = [];
    var hourAry = [];
    for (var i = 0; i < 7; i++) {
        if ((sWeek <= i && i <= eWeek) || (eWeek <= i && i <= sWeek)) {
            weekAry.push(i);
        }
    }
    for (i = 0; i < 24; i++) {
        if ((sHour <= i && i <= eHour) || (eHour <= i && i <= sHour)) {
            hourAry.push(i);
        }
    }
    allHourCache.filter('.ls').removeClass('ls');
    for (var j in weekAry) {
        for (var k in hourAry) {
            $('#week' + weekAry[j] + 'hour' + hourAry[k]).addClass('ls');
        }
    }

}