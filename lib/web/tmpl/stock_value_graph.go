package tmpl

func Stg(sma_c, sma_l, data, data2 string) string {
	return `
<!DOCTYPE html>
<html lang="en">
<html>
<head>
    <meta charset="UTF-8">
    <title>Title</title>
    <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
    <script type="text/javascript">
        google.charts.load('current', {packages: ['corechart', 'line']});
        google.charts.setOnLoadCallback(drawLineColors);

        function drawLineColors() {
            var data = new google.visualization.DataTable();
            data.addColumn('string', 'BidAt');
            data.addColumn('number', 'LastBid');
            data.addColumn('number', '` + sma_c + `');
            data.addColumn('number', '` + sma_l + `');

            data.addRows([
                ` + data + `
            ]);

            var options = {
                hAxis: {
                    title: 'Time'
                },
                vAxis: {
                    title: 'Bids'
                }
            };

            var chart = new google.visualization.LineChart(document.getElementById('chart_div'));
            chart.draw(data, options);
        }
    </script>
</head>
<body>
<table style="border-style:solid">
	<tr>
		<td><div id="chart_div" style="width: 1200px; height: 800px"></div></td>
		<td>` + data2 + `</td>
	</tr>
</table>

</body>
</html>
	`
}
