"use strict";

var socket = io();
socket.on('reply', function(msg) {
    $('#messages').append($('<li>').text(msg));
});
$('form').submit(function() {
    socket.emit('message', $('#m').val());
    $('#m').val('');
    return false;
});
