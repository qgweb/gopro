$(function() {
	function getPostData() {
		var dt = {
			app_uid: $('#app_uid').val(),
			username: $("#username").val(),
			password: $('#password').val(),
			pwd: $('#pwd').val()
		};
		return dt;
	}

	function checkPostData() {
		var pdt = getPostData();
		var err = [];
		pdt.username || err.push('邮箱不能为空');
		pdt.password || err.push('密码不能为空');
		(pdt.password.length < 6 || pdt.password.length > 20) && err.push('密码长度6~20字符');
		/^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(.[a-zA-Z0-9_-])+/.test(pdt.username) || err.push('用户名不是邮箱格式');
		(pdt.password != pdt.pwd) && err.push('两次输入的密码不相同');
		showErr(err[0]);
		return !err.length;
	}

	function doPost() {
		$.ajax({
			url: '/user/register',
			data: getPostData(),
			type: 'post',
			dataType: 'json',
			success: function(d) {
				if (d.code == "200") {
					location.href = d.redirect;
				} else if (d.code == "3xx") {
					showErr(d.msg);
				}
			}
		})
	}

	function showErr(str) {
		str && alert(str);
	}
	$('#surePost').click(function() {
		checkPostData() && doPost();
	})
	$('#regForm').submit(function() {
		return false;
	})
})