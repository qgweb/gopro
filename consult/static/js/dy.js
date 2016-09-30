$(function () {
    $('#sbtn').click(function () {
        var name = $.trim($("#name").val());
        var phone = $.trim($("#phone").val());
        var company = $.trim($("#company").val());
        var quession = $.trim($("#quession").val());
        var email = $.trim($("#email").val());

        if (!name || !phone || !company || !quession || !email) {
            layer.alert('调研信息填写不完整哦', {icon: 5});
            return;
        }
        $.ajax({
            url: "/dy",
            data: {
                name: name,
                phone: phone,
                company: company,
                quession: quession,
                email: email,
            },
            type: 'post',
            dataType: 'json',
            success: function (data) {
                if (data.ret == 0) {
                    layer.alert(data.msg, {icon: 6});
                    setTimeout("window.location.href='/'", 3000);
                } else {
                    layer.alert(data.msg, {icon: 5});
                }
            }
        })
    })
})