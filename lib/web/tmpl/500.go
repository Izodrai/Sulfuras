package tmpl


func Error500(err error) string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>ERROR 500</title>
</head>
<body>
    `+err.Error()+`
</body>
</html>
	`
}
