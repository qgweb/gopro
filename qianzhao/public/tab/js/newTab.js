window.jQuery && jQuery(function () {
    window.console || (window.console = {log: function () {
    }});
    var nowInx = $('.page.showPage').attr('inx') * 1;
    var lastPageInx = $('.page').last().attr('inx');

    function toLeft() {
        nowInx > 0 && (nowInx--, toPage(nowInx));
        return false;
    }

    function toRight() {
        nowInx < lastPageInx && (nowInx++, toPage(nowInx));
        return false;
    }


    function changeHeight(isBiger) {
        var nHei = ($(window).height() - 200) / 3;
        nHei < 95 && (nHei = 95);
        nHei > 160 && (nHei = 160);
        var nWid = nHei * (278 / 160) * 4;
        nWid > $(window).width() * 0.8 && (nWid = $(window).width() * 0.8);
        var nowWid = $('#labelsTable').width() * 1;
        if ((isBiger && (nWid >= nowWid)) || ((!isBiger) && (nWid <= nowWid))) {
            var lstbl = $('#labelsTable');
            lstbl.stop().animate({
                width: nWid + 'px'
            }, 300)
        }
    }

    function changeWidth(isBiger) {
        var nWid = ($(window).width() - 500) / 4;
        nWid < 165 && (nWid = 165);
        nWid > 278 && (nWid = 278);
        nWid * 4 > $(window).width() * 0.8 && (nWid = $(window).width() * 0.2);
        var nowWid = $('#labelsTable').width() * 0.25;
        if ((isBiger && (nWid >= nowWid)) || ((!isBiger) && (nWid <= nowWid))) {
            var lstbl = $('#labelsTable');
            lstbl.stop().animate({
                width: nWid * 4 + 'px'
            }, 300)
        }
    }

    var lastSize = {};
    var changeInvl = null;

    function changePageByResize() {
        var nowSize = {width: $(window).width() * 1, height: $(window).height() * 1};
        clearTimeout(changeInvl);
        changeInvl = setTimeout(function () {
            if (Math.abs(nowSize.width - lastSize.width) > 10) {
                (!-[1, ] && !window.XMLHttpRequest) || changeWidth(nowSize.width > lastSize.width);
            }
            if (Math.abs(nowSize.height - lastSize.height) > 10) {
                (!-[1, ] && !window.XMLHttpRequest) || changeHeight(nowSize.height > lastSize.height);
            }
            lastSize = nowSize;


            toPage(nowInx, 1);
            var $winHei = $(window).height() * 1;
            foot.css({top: ($winHei - (($winHei < (pg.offset().top * 1 + pg.height() * 1 + 100)) ? 40 : 70)) + 'px' });
        }, 200)
    }

    function toPage(inx, noAnmi) {
        $('.page').each(function (i, v) {
            var $win = $(window);
            var $widWidth = $win.width() * 1;
            var $t = $(v);
            var left = ((i - inx) * $widWidth) + 'px';

            noAnmi ? $t.css({
                left: left
            }) : $t.stop().animate({
                'left': left
            }, '700', 'linear');
        })
        $('li.clickToPage').removeClass('selected').filter('[pInx="' + inx + '"]').addClass('selected');
    }

    var shpbg = $('.shopping_bg');


    var foot = $('.footer');
    var pg = $('.page[inx=0]');

    $('.sliders_left').click(toLeft) //点击左箭头
    $('.sliders_right').click(toRight) //点击右箭头
    $('.clickToPage').click(function () {
        nowInx = $(this).attr('pInx') * 1;
        toPage(nowInx);
    })


    $(window).scroll(function () {
        var $t = $(this);
        $t.width() > 1024 && $t.scrollLeft(0);
    }).resize(function () {
        changePageByResize();
    }).trigger('resize');
    $('#labelsTable img').load(function () {
        $(window).trigger('resize');
    })

    $('#todaySure li a').click(function (e, p) { //今日值得买    1 2 3 4切换
        var $tInx = $('#todaySure li a').removeClass('selected').filter(this).addClass('selected').text() * 1 - 1;
        $('.list-container ul').each(function (i, v) {
            var $t = $(v);
            var lft = (i - $tInx) * $('#prodLst').width() + 'px';
            p == "yes" ? $t.css({
                marginLeft: lft
            }) : $t.animate({
                marginLeft: lft
            }, '300', 'linear');
        })
        return false;
    }).first().trigger('click', 'yes');

    $('#typeCheck li').mouseover(function () { //正品商家排行榜
        $('#typeCheckUls ul').hide().filter('.' + $('#typeCheck li').removeClass('selected').filter(this).addClass('selected').attr('tp')).show();
        return false;
    }).first().mouseover()

    $('body').keydown(function (e) {
        if (e.which == 37) {
            $('.sliders_left').click();
        } else if (e.which == 39) {
            $('.sliders_right').click();
        }
    })

    $('img').load(function () {
        $(window).trigger('resize');
    });

    (!-[1, ] && !window.XMLHttpRequest) && $('.sliders_left,.sliders_right').hover(function () {
        $(this).addClass('active');
    }, function () {
        $(this).removeClass('active');
    })
})