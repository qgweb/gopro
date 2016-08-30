<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="//cdn.bootcss.com/bootstrap/3.3.5/css/bootstrap.min.css">
    <script src="//cdn.bootcss.com/jquery/1.11.3/jquery.min.js"></script>
    <script src="//cdn.bootcss.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>
    <title>启冠网路</title>
</head>
<body>
<div data-example-id="togglable-tabs" style="width: 700px;margin: 0 auto;">
    <ul id="myTabs" class="nav nav-tabs" role="tablist">
        <li role="presentation" class="active"><a href="#home" id="home-tab" role="tab" data-toggle="tab" aria-controls="home" aria-expanded="true">咨询列表</a></li>
        <li role="presentation" class=""><a href="#profile" role="tab" id="profile-tab" data-toggle="tab" aria-controls="profile" aria-expanded="false">统计列表</a></li>
    </ul>
    <div id="myTabContent" class="tab-content">
        <div role="tabpanel" class="tab-pane fade active in" id="home" aria-labelledby="home-tab">
            <div>
                <table class="table" style="width:750px;">
                    <tr>
                        <td>时间</td>
                        <td>名称</td>
                        <td>电话</td>
                        <td>QQ</td>
                        <td>行业</td>
                        <td>来源</td>
                    </tr>
                    {{range $k,$d:=.list1 }}
                    <tr>
                        <td>{{$d.Date|unix}}</td>
                        <td>{{$d.Name}}</td>
                        <td>{{$d.Phone}}</td>
                        <td>{{$d.QQ}}</td>
                        <td>{{$d.Industry}}</td>
                        <td>{{$d.Reffer}}</td>
                    </tr>
                    {{end}}
                </table>
            </div>
        </div>
        <div role="tabpanel" class="tab-pane fade" id="profile" aria-labelledby="profile-tab">
            <table class="table" style="width:700px;">
                <tr>
                    <td>时间</td>
                    <td>总pv</td>
                    <td>咨询量</td>
                    <td>表单量</td>
                </tr>
                {{range $k,$d:=.list2 }}
                <tr>
                    <td>{{$d.date|unix}}</td>
                    <td>{{$d.pv}}</td>
                    <td>{{$d.cs}}</td>
                    <td>{{$d.fm}}</td>
                </tr>
                {{end}}
            </table>
        </div>
    </div>
</div>

<script>
    $('#myTabs a').click(function (e) {
        e.preventDefault()
        $(this).tab('show')
    })
</script>
</body>
</html>