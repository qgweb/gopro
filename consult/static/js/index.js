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
        var lilen = $('.example_list ul li').length - 3;
        var px = -330;
        var mleft = parseInt($('.example_list ul').css("marginLeft").replace('px', ""));
        var index = Math.floor((mleft - 10) / px);
        var bj = lilen * px - 10;
        if ((index + 1) * px + -10 < bj) {
            return;
        }
        $('.example_list ul').css("marginLeft", ((index + 1) * px + -10) + "px");
        if ((index + 1) * px + -10 <= bj) {
            $('.btn_next').hide();
        }
    });
})

$(function () {
    $('.btn_pre').click(function () {
        $('.btn_next').show();
        var px = -330;
        var mleft = parseInt($('.example_list ul').css("marginLeft").replace('px', ""));
        var index = Math.floor((mleft - 10) / px);

        if (mleft >= -10) {
            return;
        }

        $('.example_list ul').css("marginLeft", ((index - 1) * px - 10) + "px");
        if ((index - 1) * px - 10 >= -10) {
            $('.btn_pre').hide();
            return;
        }
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
                industry: industry
            },
            type: 'post',
            dataType: 'json',
            success: function (data) {
                if (data.ret == 0) {
                    $('.success').show();
                } else {
                    $('.failure').show();
                }
            }
        })


    })
})
//统计
$(function () {
    $("body").append('<img src="/sts?t=3"/>');
    $("#talkByQQh1,#talkByQQh2,#talkByQQh3").click(function () {
        $("body").append('<img src="/sts?t=1"/>');
    })
})
$(function(){
    function changeIMGWidth() {
        var w=$(window).width()/2+94;
        $('.banner').width(w);
        $('.banner li').width(w);
    }
    $(window).resize(function() {
        changeIMGWidth();
    })
    changeIMGWidth();
})