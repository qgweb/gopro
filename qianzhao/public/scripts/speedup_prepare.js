$(function () {
    function clickBtn() {//点击开关
        var $t = $('.on_off');

        $t.hasClass('btn_on') || $.ajax({
            url: '/operate/speedup_open',
            type: 'post',
            dataType: 'json',
            data: {sid: $('body').data('sid')},
            success: function (d) {
                if (d && d.code == '200') {
                    $t.addClass('btn_on');
                    setDjs(d.start_time * 1000, 3600000);
                } else if (d) {
                    $t.removeClass('btn_on');
                    $('span.msgCnt').remove();
                    $('#djsSpan').before('<span class="msgCnt">' + d.msg + '</span>');
                }
            }
        })
    }

    var cj = 0;
    //开始时间,持续时间
    function setDjs(start, line) {//倒计时
        $('span.msgCnt').remove();
        $('#djsSpan').before('<span class="msgCnt">您好，当前剩余加速时间</span>');
        $('.high_speed_mode').show();
        cj = start - (new Date());
        $('#djsSpan').trigger('startDjs', {start: start, line: line});
    }

    function getRealTime() {
        return (new Date()) * 1 + cj;
    }

    function numaddZero(num) {
        if (num * 1 < 10) {
            num = '0' + num;
        }
        return num + '';
    }

    var invl = null;
    $('#djsSpan').on('startDjs', function (e, p) {
        var $t = $(this);
        clearInterval(invl);
        invl = setInterval(function () {
            var hm = getRealTime() - p.start;//已经过了的时间
            if (p.line < hm) {//倒计时 时间到
                $('.msgCnt').text('抱歉，您今天的免费加速时间已到，请明天再来~');
                $t.text('');
            } else {
                var tmx = p.line - hm;//剩下的时间
                var mnt = Math.floor(tmx / 60000);//分
                var sec = Math.floor((tmx - mnt * 60000) / 1000);//秒
                $t.text(numaddZero(mnt) + ':' + numaddZero(sec));
            }
        }, 300);
    })
    $.ajax({
        url: '/operate/speedup_open_check',
        type: 'post',
        dataType: 'json',
        success: function (d) {
            if (d && d.code != "200") {
                $('span.msgCnt').remove();
                $('#djsSpan').before('<span class="msgCnt">' + d.message + '</span>');
            } else if (d && d.code == "200") {//
                $('span.msgCnt').remove();
                $('#djsSpan').before('<span class="msgCnt">恭喜，您的运行环境符合加速条件，请点击按钮进行免费提速</span>');
                $('body').data('sid', d.sid);
                $('.on_off').click(clickBtn);
                if (d && d.start_time && d.cur_time) {
                    $('.on_off').addClass('btn_on');
                    setDjs(d.start_time * 1000, d.start_time * 1000 + 3600000 - d.cur_time * 1000);
                }
            }
        }
    });

});