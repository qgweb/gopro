//营销qq
$(function () {
    BizQQWPA.addCustom({
        aty: '0', //接入到指定工号
        nameAccount: "800066774", //营销QQ号码
        selector: "talkByQQh3",
    });
    BizQQWPA.addCustom({
        aty: '0', //接入到指定工号
        nameAccount: "800066774", //营销QQ号码
        selector: "talkByQQh1",
    });
    BizQQWPA.addCustom({
        aty: '0', //接入到指定工号
        nameAccount: "800066774", //营销QQ号码
        selector: "talkByQQh2",
    });
})
//点击移动
$(function () {
    $('.btn_next').click(function () {
        $('.btn_pre').show();
        var curLeft = parseInt($('.example_list ul').css("marginLeft").replace('px', "")) - 335;
        if (curLeft <= -1015) {
            return;
        }
        if (curLeft == -680) {
            $('.btn_next').hide();
        }
        $(".example_list ul").animate({marginLeft: curLeft + "px"}, "slow");
    });
})

$(function () {
    $('.btn_pre').click(function () {
        $('.btn_next').show();
        var curLeft = parseInt($('.example_list ul').css("marginLeft").replace('px', "")) + 335;
        if (curLeft >= 325) {
            return;
        }
        if (curLeft == -10) {
            $('.btn_pre').hide();
        }
        $(".example_list ul").animate({marginLeft: curLeft + "px"}, "slow");
    });
})
//表单提交
$(function () {
    $('.success input,.failure input').click(function () {
        $(this).parents("div").hide();
    })

    $('.form_btn').click(function () {
        var name = $.trim($("#fm_name").val());
        var phone = $.trim($("#fm_phone").val());
        var qq = $.trim($("#fm_qq").val());
        var industry = $.trim($("#fm_industry").val());
        var reffer = window.location.href;

        if (!name || !phone || !qq || !industry) {
            layer.alert('预约信息填写不完整哦', {icon: 5});
            return;
        }
        $("body").append('<img src="/sts?t=2"/>');
        $.ajax({
            url: "/submit",
            data: {
                name: name,
                phone: phone,
                qq: qq,
                industry: industry,
                reffer: reffer,
            },
            type: 'post',
            dataType: 'json',
            success: function (data) {
                if (data.ret == 0) {
                    $('.success').show();
                    setTimeout("window.location.reload()", 1000);
                } else {
                    $('.failure').show();
                }
            }
        })


    })
})
//统计
$(function () {
    $("body").append('<img src="/sts?t=3" style="display:none;"/>');
    $("#talkByQQh1,#talkByQQh2,#talkByQQh3").click(function () {
        $("body").append('<img src="/sts?t=1" style="display:none;"/>');
    })
})
$(function () {
    function changeIMGWidth() {
        var w = $(window).width() / 2 + 94;
        $('.banner').width(w);
        $('.banner li').width(w);
    }

    $(window).resize(function () {
        changeIMGWidth();
    })
    changeIMGWidth();
})