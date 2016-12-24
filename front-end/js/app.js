
$(function () {
    var params = {name: "EOH"};
    $.getJSON("getstockprices", params, function (data) {
        console.info(data);

        Highcharts.stockChart("container", {

            rangeSelector: {
                selected: 1
            },

            title: {
                text: 'EOH Stock Price'
            },

            series: [{
                name: params.name,
                data: data,
                type: "area",
                threshold: null,
                tooltip: {
                    valueDecimals: 2
                },
                fillfactor: {
                    linearGradient: {
                        x1: 0,
                        y1: 0,
                        x2: 0,
                        y2: 1
                    }
                }
            }]
        });
    });
});
