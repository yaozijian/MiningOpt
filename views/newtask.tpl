<!DOCTYPE html>
<html lang="us-EN">
<head>
	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="HandheldFriendly" content="True" />
	<link rel="shortcut icon" href="/static/favicon.ico">
	<title>New Task Result</title>

	<script type="text/javascript">
		function countDown(secs,surl){
			var item = document.getElementById('seconds_left');
			item.innerHTML = secs;
			if(--secs > 0){
				setTimeout("countDown(" + secs + ",'" + surl + "')", 1000);
			}else{
				location.href = surl;
			}
		}
	</script>
</head>
<body>
	<h1>
	{{if lt 0 (.msg | len ) }}
		{{print .msg}}
	{{else}}
		{{print .error}}
	{{end}}
	</h1>
	<br>
	Go to main page after <span id="seconds_left">8</span> seconds.
	<script type="text/javascript">countDown(3,"/");</script>
</body>
</html>