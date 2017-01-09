var reportName = "sewer-gravitypipebreakdown-bymaterial";

function pieChart(id, title, seriesName, seriesData) {
    Highcharts.chart(id, {
        chart: {
            plotBackgroundColor: null,
            plotBorderWidth: null,
            plotShadow: false,
            type: "pie"
        },
        title: {
            text: title
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
            name: seriesName,
            colorByPoint: true,
            data: seriesData
        }]
    });
}

function columnChart(id, title, xAxisTitle, xAxisCategories, yAxisTitle, series) {
    Highcharts.chart(id, {
        chart: {
            type: "column"
        },
        title: {
            text: title
        },
        xAxis: {
            title: {
                text: xAxisTitle
            },
            categories: xAxisCategories,
            crosshair: true
        },
        yAxis: {
            min: 0,
            title: {
                text: yAxisTitle
            }
        },
        series: series
    });
}

$(function () {
    $.getJSON(reportName + "/data", null, function (data) {

        var piedata_length = [];
        var piedata_replacevalue = [];
        for (var i = 0; i < data["rows"].length; i++) {
            var slice_length = {"name": data["rows"][i][0]};
            var slice_replacevalue = {"name": data["rows"][i][0]};
            var length = parseFloat(data["rows"][i][2]);
            var replacevalue = parseFloat(data["rows"][i][3]);
            slice_length["y"] = isNaN(length) ? 0 : length;
            slice_replacevalue["y"] = isNaN(replacevalue) ? 0 : replacevalue;
            piedata_length.push(slice_length);
            piedata_replacevalue.push(slice_replacevalue)
        }

        pieChart("chart-pie-length", "Length (m) per Material", "Material", piedata_length);
        pieChart("chart-pie-replacevalue", "Replace Value (R) per Material", "Material", piedata_replacevalue);

        var columnCategories = [];
        var columndata_length = {"name": "Material", "data": []};
        var columndata_replacevalue = {"name": "Replace Value (R)", "data": []};
        for (var i = 0; i < data["rows"].length; i++) {
            columnCategories.push(data["rows"][i][0]);
            var length = parseFloat(data["rows"][i][2]);
            var replacevalue = parseFloat(data["rows"][i][3]);
            columndata_length.data.push(isNaN(length) ? 0 : length);
            columndata_replacevalue.data.push(isNaN(replacevalue) ? 0 : replacevalue)
        }

        columnChart("chart-column-length", "Length (m) per Material", "Material", columnCategories, "Length (m)", [columndata_length]);
        columnChart("chart-column-replacevalue", "Replace Value (R) of Material", "Material", columnCategories, "Replace Value (R)", [columndata_replacevalue]);

        var gridColumns = [];
        var gridRows = [];

        var options = {
            enableCellNavigation: true,
            enableColumnReorder: false
        };

        var columns = data["cols"] || [];
        for (var i = 0; i < columns.length; i++) {
            var col = {id: columns[i], name: columns[i], field: columns[i]};
            if (col.name.indexOf("Replace Value") !== -1) // TODO: make this better
                col["formatter"] = Slick.Formatters.CurrencyRand;

            gridColumns.push(col);
        }
        for (var i = 1; i < columns.length; i++) {
            gridColumns[i]["hasTotal"] = true;
        }

        // Rows
        var rows = data["rows"] || [];
        for (var i = 0; i < rows.length; i++) {
            var record = {};
            for (var j = 0; j < columns.length; j++)
                record[columns[j]] = rows[i][j];

            gridRows.push(record);
        }
        if (gridRows.length > 0) {
            var dataProvider = new TotalsDataProvider(gridRows, gridColumns);
            var grid = new Slick.Grid("#table-summary", dataProvider, gridColumns, options);
            grid.autosizeColumns();
        }
    });
});

function TotalsDataProvider(data, columns) {
    var totals = {};
    var totalsMetadata = {
        // Style the totals row differently.
        cssClasses: "totals",
        columns: {}
    };

    // Make the totals not editable.
    for (var i = 0; i < columns.length; i++) {
        totalsMetadata.columns[i] = { editor: null };
    }


    this.getLength = function() {
        return data.length + 1;
    };

    this.getItem = function(index) {
        return (index < data.length) ? data[index] : totals;
    };

    this.updateTotals = function() {
        var columnIdx = columns.length;
        while (columnIdx--) {
            var columnId = columns[columnIdx].id;
            var total = 0;
            var i = data.length;
            while (i--) {
                total += (parseInt(data[i][columnId], 10) || 0);
            }
            totals[columnId] = "Sum:  " + total;
        }
    };

    this.getItemMetadata = function(index) {
        return (index != data.length) ? null : totalsMetadata;
    };

    this.updateTotals();
}
