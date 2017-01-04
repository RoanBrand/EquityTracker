var reply;

$(function () {
    $.getJSON("report-water-modelsummaries-byelements", null, function (data) {
        console.info(data); reply = data;

        Highcharts.chart("columnchart", {
            chart: {
                type: "column"
            },
            title: {
                text: "Report - Water - Model Summaries - by Elements"
            },
            xAxis: {
                title: {
                    text: "Elements"
                },
                categories: data["Categories"],
                crosshair: true
            },
            yAxis: {
                min: 0,
                title: {
                    text: "Replace Value (R)"
                }
            },
            series: data["Series"]
        });

        var piedata = [];
        for (var i = 0; i < data["Categories"].length; i++) {
            piedata.push({"name": data["Categories"][i], "y": data["Series"][0]["data"][i]})
        }
        console.info(piedata);
        Highcharts.chart("piechart", {
            chart: {
                plotBackgroundColor: null,
                plotBorderWidth: null,
                plotShadow: false,
                type: "pie"
            },
            title: {
                text: "Report - Water - Model Summaries - by Elements"
            },
            plotOptions: {
                pie: {
                    allowPointSelect: true,
                    cursor: 'pointer',
                    dataLabels: {
                        enabled: true,
                        format: '<b>{point.name}</b>: {point.percentage:.1f} %',
                        style: {
                            color: (Highcharts.theme && Highcharts.theme.contrastTextColor) || 'black'
                        }
                    }
                }
            },
            series: [{
                name: "Water pie chart",
                colorByPoint: true,
                data: piedata
            }]
        });

        var gridColumns = [];
        var gridRows = [];

        var options = {
            enableCellNavigation: true,
            enableColumnReorder: false
        };

        $.getJSON("view-Water-ModelSummariesbyElements", null, function(res) {
            console.info(res);
            var columns = res["cols"] || [];
            for (var i = 0; i < columns.length; i++) {
                gridColumns.push({id: columns[i], name: columns[i], field: columns[i]});
            }
            var rows = res["rows"] || [];
            for (var i = 0; i < rows.length; i++) {
                var record = {};
                for (var j = 0; j < columns.length; j++) {
                    record[columns[j]] = rows[i][j];
                }
                gridRows.push(record);
            }

            var grid = new Slick.Grid("#myGrid", gridRows, gridColumns, options);
        });

    });
});
