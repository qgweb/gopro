$(function() {
	function getPostData() {
		var dt = {
			username: $("#username").val(),
			password: $('#password').val(),
			remember: $('#remember')[0].checked ? 1 : 0
		};
		return dt;
	}

	function checkPostData() {
		var pdt = getPostData();
		var err = [];
		pdt.username || err.push('用户名不能为空');
		pdt.password || err.push('密码不能为空');
		(pdt.password.length < 6 || pdt.password.length > 20) && err.push('密码错误');
		/^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(.[a-zA-Z0-9_-])+/.test(pdt.username) || err.push('用户名错误');
		showErr(err[0]);
		return !err.length;
	}

	function doPost() {
		$.ajax({
			url: '/user/login',
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
		str && $('#errShow').text(str).show();
	}
	$('#surePost').click(function() {
		checkPostData() && doPost();
	})
	$('#logForm').submit(function() {
		return false;
	})
})