$(function () {
    $('body').on('click', 'a[href="#"]', function (e) {
        e.preventDefault();
    })
    var bDown = true;
    var lastAngel = null;

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
                bDown = true;
            }
        })
    }

    $('.clickRotate').on('click', function () {
        var ran = Math.random();
        $.ajax({
            url: '/club/turntable',
            type: 'get',
            dataType: 'json',
            success: function (d) {
                if (bDown) {
                    bDown = false;
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
                        var coins = ['5', '1', '0', '10', '5', '1', '0', '10'][resl];
                        rotateFunc($('.arrow'), 45 * resl + 22.5, function () {
                            $('.popup_bg').show()
                            $('.popup').show();
                            if (html == "谢谢参与") {
                                $('.win_lottery .c_text').html('<p class="c_text"><span>' + html + '</span>！</p>')
                                $('.win_lottery .pwd,.win_lottery .plink').hide();
                                //alert('谢谢参与');
                                $('.win_lottery .c_text').eq(0).css({
                                    marginBottom: 10
                                })
                                $('.win_lottery .content').css({
                                    height: 147
                                })
                                $('.win_lottery .popupc_bg').css({
                                    height: 158
                                })
                            } else {
                                $('.win_lottery .c_text').eq(0).css({
                                    marginBottom: 0
                                })
                                $('.win_lottery .content').css({
                                    height: 226
                                })
                                $('.win_lottery .popupc_bg').css({
                                    height: 237
                                })
                                $('.win_lottery .c_text').html('<p class="c_text">恭喜你！抽中<span>' + html + '</span>！</p>')
                                $('.win_lottery .pwd,.win_lottery .plink').show();
                                $('.win_lottery .pwd p').eq(1).html(d.rcode);
                            }
                            $('.win_lottery').show();
                        })
                    } else if (d.code == 301) {
                        if (d.msg == "您今天抽奖次数已经用完，请明天再来！") {
                            $('.prompt').show()
                        } else {
                            $('.login_prompt').show();
                        }
                        $('.popup_bg').show();
                        $('.popup').show()
                        bDown = true;
                    } else {
                        alert(d.msg);
                        bDown = true;
                    }
                    //console.log(bDown);
                    //bDown=true
                }
            }

        })
    })
    //获奖记录滚动
    var timer = null;

    function move() {
        var top = $('.winning_record ul').height() / 2;
        var start = {};
        var dis = {};
        start['top'] = parseInt($('.winning_record ul').css('top')) || 0;
        n = 0;
        timer = setInterval(function () {
            n += 1;
            var cur = start['top'] - n;
            $('.winning_record ul').css({top: cur});
            if (Math.abs(cur) == top) {
                clearInterval(timer)
                $('.winning_record ul').css({top: 0})
                move();
            }
        }, 30)
    }

    function roundList() {
        $('.winning_record ul').html($('.winning_record ul').html() + $('.winning_record ul').html())
        if ($('.winning_record ul li').length > 14) {
            move()
        }
    }

    function getList() {
        $.ajax({
            url: '/club/winlist',
            type: 'get',
            dataType: 'json',
            success: function (d) {
                var html = '';
                for (var i = 0; i < d.length; i++) {
                    html += '<li><span>' + d[i].Phone + '</span><span>' + d[i].Result + '</span><span>' + d[i].Date + '</span></li>'
                }
                $('.winning_record ul').html(html);
                roundList()
            }
        })
    }

    getList();
    //中奖记录
    $('.myWin').on('click', function () {
        $.ajax({
            url: '/club/mywin',
            dataType: 'json',
            type: 'get',
            success: function (d) {
                if (d.msg) {
                    $('.login_prompt').show();
                    $('.popup_bg').show();
                    $('.popup').show()
                } else {
                    $('.my_record .pwd .myWinList').remove();
                    var list = '';
                    for (var i = 0; i < d.length; i++) {
                        list += '<p class="myWinList">' + d[i].code + '</p>';
                    }
                    $('.my_record .pwd').html(list);
                    $('.my_record').show();
                    $('.popup_bg').show()
                    $('.popup').show()
                }

            }
        })
    })
    //关闭
    $('.btn_close').on('click', function () {
        $('.popup_bg').hide();
        $('.popup').hide();
        $(this).parents('.popup_c').hide();
    })
})



