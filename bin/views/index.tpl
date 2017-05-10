<!DOCTYPE html>
<html lang="us-EN">
<head>
	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="HandheldFriendly" content="True" />
	<link rel="shortcut icon" href="/static/favicon.ico">	
	<title>Mining Optimization</title>
    <!-- Bootstrap -->
    <script src="static/jquery-1.12.4/jquery.min.js"></script>
    <script src="static/bootstrap-3.3.7/js/bootstrap.min.js"></script>
    <link href="static/bootstrap-3.3.7/css/bootstrap.min.css" rel="stylesheet">
    <!-- Bootstrap file input from http://plugins.krajee.com/file-input --> 
    <link href="static/bootstrap-fileinput-4.3.9/css/fileinput.css" media="all" rel="stylesheet" type="text/css"/>    
    <script src="static/bootstrap-fileinput-4.3.9/js/fileinput.js" type="text/javascript"></script>
    <script src="static/bootstrap-fileinput-4.3.9/js/locales/es.js" type="text/javascript"></script>
    <script src="static/bootstrap-fileinput-4.3.9/themes/fa/theme.js" type="text/javascript"></script>
    <script type="text/javascript">
    	$(document).on('ready',function() {
    		var cfg = {
    			"showUpload": false,
    			"maxFileCount": 1,
    		};    		
    		$("#data-file").fileinput(cfg);
    		$("#param-file").fileinput(cfg);
    	});
    </script>
</head>
<body>
<div class="container-fluid">
	<div class="row-fluid">
		<div class="span12">
			<div class="tabbable" id="tabs-main">
				<ul class="nav nav-tabs">
					<li class="active">
						<a href="#panel-task-list" data-toggle="tab">Task List</a>
					</li>
					<li>
						<a href="#panel-server-list" data-toggle="tab">Server List</a>
					</li>
					<li>
						<a href="#panel-new-task" data-toggle="tab">New Task</a>
					</li>
				</ul>
				<div class="tab-content">
					<div class="tab-pane active" id="panel-task-list">
						<table class="table">
							<thead>
								<tr>
									<th>ID</th>
									<th>Status</th>
									<th>Server</th>
									<th>Result</th>
									<th>Create At</th>
									<th>Data</th>
									<th>Param</th>
								</tr>
							</thead>
							<tbody>
								{{range .tasklist}}
									<tr>
										<td>{{.Id}}</td>
										<td>{{.Status}}</td>
										<td></td>
										<td></td>
										<td>{{.Create}}</td>
										<td><a href="{{.DataURL}}">data.gz</a></td>
										<td><a href="{{.ParamURL}}">param.json</a></td>
									</tr>
								{{end}}
							</tbody>
						</table>
					</div>
					<div class="tab-pane" id="panel-server-list">
						<table class="table">
							<thead>
								<tr>
									<th>Address</th>
									<th>Task ID</th>
									<th>Task Progress</th>
									<th>Online At</th>
								</tr>
							</thead>
							<tbody>
								{{range .serverlist}}
									<tr>
										<td>{{.Address}}</td>
										<td></td>
										<td></td>
										<td>{{.OnlineAt}}</td>
									</tr>
								{{end}}
							</tbody>
						</table>
					</div>
					<div class="tab-pane" id="panel-new-task">
						<form class="form-new-task" enctype="multipart/form-data" action="/" method="post">
							<label class="control-label">Data File</label>
							<input id="data-file" name="data-file" type="file" class="file-loading" data-allowed-file-extensions='["gz","tar.gz","tgz"]'/>
							<label class="control-label">Param File</label>
							<input id="param-file" name="param-file" type="file" class="file-loading" data-allowed-file-extensions='["json"]'/><br>
							<center><button class="btn btn-large btn-primary" type="submit">New Task</button></center>
						</form>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
</body>
</html>