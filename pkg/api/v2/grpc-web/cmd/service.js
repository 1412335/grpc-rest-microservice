const app = {};

app.Service = function (service, ctors) {
    this.service = service
    this.ctors = ctors
    this.textEncoded = true
};

app.Service.INTERVAL = 500;
app.Service.MAX_STREAM_MESSAGES = 10;

app.Service.addMessage = function (message, cssClass) {
    console.log(`add message "${message}" with css class "${cssClass}"`)
    $("#first").after(
        $("<div/>").addClass("row").append(
            $("<h2/>").append(
                $("<span/>").addClass("label label-" + cssClass).text(message))));
}

app.Service.prototype.ping = function (timestamp) {
    var unaryRequest = new this.ctors.MessagePing()
    unaryRequest.setTimestamp(timestamp)

    var metadata = { 'custom-header-1': 'value1' }
    var call = (this.textEncoded ? this.service.clientText : this.service.clientBinary).ping(unaryRequest, metadata, function (err, response) {
        if (err) {
            app.Service.addMessage(`Error code: ${err.code} "${err.message}"`, 'danger')
        } else {
            setTimeout(() => {
                app.Service.addMessage(`Resp: ${response.getTimestamp()} - ${response.getServicename()}`, 'success')
            }, app.Service.INTERVAL);
        }
    })

    call.on('status', function (status) {
        if (status.metadata) {
            console.log('Received metadata')
            console.log(status.metadata)
        }
    })
}

// app.Service.prototype.pingError = function (timestamp) {
//     var unaryRequest = new this.ctors.MessagePing()
//     unaryRequest.setTimestamp(timestamp)
//     this.service.pingAbort(unaryRequest, {}, function (err, response) {
//         if (err) {
//             app.Service.addMessage(`Error code: ${err.code} "${err.message}"`, 'danger')
//         }
//     })
// }

app.Service.prototype.repeatPing = function (timestamp, count) {
    if (count > app.Service.MAX_STREAM_MESSAGES) {
        count = app.Service.MAX_STREAM_MESSAGES;
    }
    var streamRequest = new this.ctors.StreamingMessagePing();
    streamRequest.setTimestamp(timestamp);
    streamRequest.setMessageCount(count);
    streamRequest.setMessageInterval(app.Service.INTERVAL);

    var metadata = { 'custom-header-1': 'value2' }
    var stream = (this.textEncoded ? this.service.clientExtra : this.service.clientExtraBinary).streamingPing(
        streamRequest,
        metadata,
    );
    stream.on('data', function (response) {
        app.Service.addMessage(`Resp: ${response.getTimestamp()} - ${response.getServicename()}`, 'success')
    })
    stream.on('status', function (status) {
        if (status.metadata) {
            console.log('Received metadata')
            console.log(status.metadata)
        }
    })
    stream.on('error', function (err) {
        app.Service.addMessage(`Error code: ${err.code} "${err.message}"`, 'danger')
    })
    stream.on('end', function () {
        console.log('Stream end signal received')
    })
}

app.Service.prototype.send = function (e) {
    var msg = $('#msg').val().trim()
    $('#msg').val('')
    if (!msg) return false
    if (msg.indexOf(' ') > 0) {
        var count = msg.substr(0, msg.indexOf(' '))
        if (/^\d+$/.test(count)) {
            this.repeatPing(msg.substr(msg.indexOf(' ') + 1), count)
        }
    } else if (/^\d+$/.test(msg)) {
        this.ping(msg)
    } else {
        app.Service.addMessage(`Error: timestamp invalid format`, 'danger')
    }
    return false
}

app.Service.prototype.load = function () {
    var self = this
    $(document).ready(function () {
        $('#send').click(function (e) {
            self.textEncoded = true
            self.send(e)
        })
        $('#send-binary').click(function (e) {
            self.textEncoded = false
            self.send(e)
        })
        $('#msg').keyup(function (e) {
            if (e.keyCode == 13) self.send()
            return false
        })
        return false
    })
}

module.exports = app