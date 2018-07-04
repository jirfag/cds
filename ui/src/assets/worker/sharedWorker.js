importScripts('./common.js');
importScripts('./eventsource.js');

var sse;
const connections = [];

onconnect = function(e) {
    var port = e.ports[0];
    connections.push(port);
    port.onmessage = function (event) {
        if (!sse && event.data.sseURL) {
            sse = connectSSE(event.data.sseURL, event.data.headAuthKey, event.data.headAuthValue);
            sse.onmessage = function(evt) {
                // if ack get UUID
                if (evt.data.indexOf('ACK: ') === 0) {
                    return;
                }
                let jsonEvent = JSON.parse(evt.data);

                connections.forEach(p => {
                   p.postMessage(jsonEvent);
                });
                return;
            };
        }
    };
};