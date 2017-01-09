var reportName = "sewer-gravitypipebreakdown-bymaterial";

$(function () {
    $.getJSON(reportName + "/data", null, function (data) {

        var piedata = [];
        for (var i = 0; i < data["rows"].length; i++) {
            var data = {"name": data["rows"][i][0]};
            var num = parseFloat(data["rows"][i][1]);
            data["y"] = isNaN(num) ? 0 : num;
            piedata.push(data);
        }

        Highcharts.chart("chart-pie-length", {
            chart: {
                plotBackgroundColor: null,
                plotBorderWidth: null,
                plotShadow: false,
                type: "pie"
            },
            title: {
                text: "Report - Sewer - Gravity Pipe Breakdown - by Material"
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
                name: "Elements",
                colorByPoint: true,
                data: piedata
            }]
        });

/*

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
                categories: categories,
                crosshair: true
            },
            yAxis: {
                min: 0,
                title: {
                    text: "Replace Value (R)"
                }
            },
            series: [columnChartData]
        });



        var gridColumns = [];
        var gridRowsLinks = [];
        var gridRowsNodes = [];
        var gridRowsOther = [];

        var options = {
            enableCellNavigation: true,
            enableColumnReorder: false
        };

        var columns = data["cols"] || [];
        for (var i = 0; i < columns.length; i++) {
            var col = {id: columns[i], name: columns[i], field: columns[i]};
            if (col.name.indexOf("(R)") !== -1)
                col["formatter"] = Slick.Formatters.CurrencyRand;
            gridColumns.push(col);
        }

        // Rows
        var rows = data["rows"] || [];
        for (var i = 0; i < rows.length; i++) {
            var record = {};
            for (var j = 0; j < columns.length; j++) {
                record[columns[j]] = rows[i][j];
            }
            switch (record["Elements"]) {
                case "PIPE":
                case "CV":
                case "Pump":
                case "Valve (PRV)":
                case "Valve (FCV)":
                case "Valve (PSV)":
                case "Valve (PBV)":
                case "Valve (TCV)":
                    gridRowsLinks.push(record);
                    break;
                case "GL_Tank":
                case "Tower":
                case "Tank":
                case "BPT":
                case "Bulk":
                case "WTP":
                case "Well":
                case "BoreHole":
                case "Dam":
                case "River":
                case "Node":
                    gridRowsNodes.push(record);
                    break;
                default:
                    gridRowsOther.push(record);
            }
        }
        if (gridRowsLinks.length > 0) {
            (new Slick.Grid("#table-links", gridRowsLinks, gridColumns, options)).autosizeColumns();
        }
        if (gridRowsNodes.length > 0) {
            (new Slick.Grid("#table-nodes", gridRowsNodes, gridColumns, options)).autosizeColumns();
        }
        if (gridRowsOther.length > 0) {
            (new Slick.Grid("#table-other", gridRowsOther, gridColumns, options)).autosizeColumns();
        }*/
    });
});
