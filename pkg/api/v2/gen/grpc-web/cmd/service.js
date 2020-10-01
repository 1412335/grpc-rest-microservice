const app = {};

app.Service = function (service, ctors) {
    this.service = service
    this.ctors = ctors
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
    var call = this.service.ping(unaryRequest, metadata, function (err, response) {
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

app.Service.prototype.send = function (e) {
    var msg = $('#msg').val().trim()
    $('#msg').val('')
    if (!msg) return false
    if (msg.indexOf(' ') > 0) {

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
        $('#send').click(self.send.bind(self))
        $('#msg').keyup(function (e) {
            if (e.keyCode == 13) self.send()
            return false
        })
        return false
    })
}

module.exports = app