var obj = {
    Email: '',
    Username: '',
    Message: '',
    Event: '',
    Room: '',
    Server: ''
};

var conn; // conexao websocket
var chat = document.getElementById("chat");
var roomList = document.getElementById("Salas");

function newRoom() {
    //criar uma nova sala
    var roomNew = document.getElementById("inputNewRoom").value;
    obj.Event = 'add';
    obj.Message = "adicionar sala";
    obj.Room = roomNew;
    sendMsg(obj);
    console.log('New Room!');
    $('#NewRoom').modal('hide');
};

function joinRoom() {
    var chat = JSON.parse(localStorage.getItem('objChat'));
    chat.Room = document.getElementById("inputRoom").value;
    localStorage.setItem('objChat', JSON.stringify(chat));
    //enviar msg para join new room
    obj.Event = 'join';
    obj.Room = chat.Room;
    obj.Message = "entrar em sala";
    sendMsg(obj);
    console.log('Join Room!');
    $('#joinRoom').modal('hide');
};

function leaveRoom(){
    chat.Room = document.getElementById("inputLeaveRoom").value;
    localStorage.removeItem('objChat');
    //enviar msg para join new room
    obj.Event = 'leave';
    obj.Room = chat.Room;
    obj.Message = "sair da sala";
    sendMsg(obj);
    console.log('Leave Room!');
    $('#leaveRoom').modal('hide');
}

function enterChat() {
    var chat = {
        email: document.getElementById("inputEmail").value,
        username: document.getElementById("inputUser").value,
        room: 'root',
    };
    var rooms = [];

    localStorage.setItem('objChat', JSON.stringify(chat));
    localStorage.setItem('rooms', JSON.stringify(rooms));
    obj.Username = chat.username;
    obj.Message = "mudar nome";
    obj.Email = chat.email;
    obj.Event = 'change';
    sendMsg(obj);
    setTimeout(func => {
        obj.Message = "adicionar sala";
        obj.Room = 'root';
        obj.Event = 'add';
        sendMsg(obj)
    }, 3000);
    
    //console.log('Enter Chat!');
    $('#enter').modal('hide');
};

function changeUsername() {
    var chat = JSON.parse(localStorage.getItem('objChat'));
    chat.Username = document.getElementById("inputUsername").value;
    localStorage.setItem('objChat', JSON.stringify(chat));
    //enviar msg para join new room
    obj.Event = 'change';
    obj.Username = chat.Username;
    obj.Message = "mudar nome";
    sendMsg(obj);
    console.log('Change Username!');
    $('#changeNick').modal('hide');
};

function abrirPopup(n){
    if(n == 1){
        $('#NewRoom').modal('show')
    } else if (n == 2){
        $('#joinRoom').modal('show')
    } else if (n ==3){
        $('#changeNick').modal('show')
    } else if (n ==4){
        $('#LeaveRoom').modal('show')
    }
}

function sendMsg(msg) {
    if (!conn) {
        return false;
    }
    console.log('Enviando msg: ', msg)
    conn.send(
        JSON.stringify(msg)
    );
    return true;
};

function appendChat(m){
    var div = document.createElement("div");
    div.className = "media text-muted pt-3";
    var img = document.createElement("img");
    img.src = gravatar(m.email);
    img.className = "bd-placeholder-img mr-2 rounded";
    img.width = 32;
    img.height = 32;
    var p = document.createElement("p");
    p.className = "media-body pb-3 mb-0 small lh-125 border-bottom border-gray";
    var strong = document.createElement("strong");
    strong.className = "d-block text-gray-dark";
    strong.innerHTML = "@"+m.username;
    p.appendChild(strong);
    var textnode = document.createTextNode(m.message);
    p.appendChild(textnode);
    div.appendChild(img);
    div.appendChild(p);
    chat.appendChild(div)
}


function appendRoom(room){
    var item = document.createElement("li");
    var a = document.createElement("a");
    a.href = "javascript:abrirSala('"+room+"')";
    a.innerText = 'Sala #'+room;
    item.appendChild(a);
    roomList.appendChild(item);
}

function gravatar(email) {
    return 'http://www.gravatar.com/avatar/' + md5(email);
}

function addMsg(msg){
    var msgs = JSON.parse(localStorage.getItem(msg.room));
    if(msgs == null){
        msgs = [];
    }
    msgs.push(msg);
    console.log(msgs)
    localStorage.setItem(msg.room, JSON.stringify(msgs));
}

function addSala(sala){
    var salas = JSON.parse(localStorage.getItem("rooms"));
    if(salas == null){
        salas = [];
    }
    s = salas.find(function(element) {
        return element == sala;
    })
    if(s == null){
        //posso criar a sala
        salas.push(sala);
        appendRoom(sala);
    }
    localStorage.setItem("rooms", JSON.stringify(salas));
}

function abrirSala(sala){
    obj.Room = sala;
    c = chat.children;
    console.log(c);
    for (i = 0; i < c.length; i++) {
        if(c[i].nodeName == "DIV"){
            chat.removeChild(c[i]);
        }
    }
    //adicionar msg das salas
    /*var msgs = JSON.parse(localStorage.getItem(sala));
    if(msgs == null){
        msgs = [];
    }
    msgs.forEach(element => {
        appendChat(element);
    });*/
}

function Enviar(){
    if (!conn) {
        return false;
    }
    var chat = JSON.parse(localStorage.getItem('objChat'));
    obj.Message = document.getElementById("msg").value;
    obj.Event = "msg";
    sendMsg(obj);
    document.getElementById("msg").value = "";
    console.log('Msg Enviada!')
    return true;
}

window.onload = function () {
    $('#enter').modal('show')
    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws");
        conn.onclose = function (evt) {
            console.log("Connection closed - ", evt)
            //disparar reconexão com o socket
        };
        conn.onmessage = function (evt) {
            msgEvt = JSON.parse(evt.data)
            if(msgEvt.event == "msg"){
                appendChat(msgEvt);
                addMsg(msgEvt);
            } else if(msgEvt.event == "command" && msgEvt.message == "add sala"){
                //adicionar sala ao array de salas
                addSala(msgEvt.room)
            } else {
                // tratar outros tipos de mensagens
                console.log(msgEvt)
            }
        };
    } else {
        console.log("Your browser does not support WebSockets.")
    }
}