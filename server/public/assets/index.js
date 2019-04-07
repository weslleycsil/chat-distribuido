var obj = {
    Email: '',
    Username: '',
    Message: '',
    Event: '',
    Room: ''
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

function enterChat() {
    var chat = {
        email: document.getElementById("inputEmail").value,
        username: document.getElementById("inputUser").value,
        room: 'root',
    };
    localStorage.setItem('objChat', JSON.stringify(chat));
    obj.Username = chat.username;
    obj.Message = "mudar nome";
    obj.Email = chat.email;
    obj.Event = 'change';
    sendMsg(obj);
    obj.Message = "adicionar sala";
    obj.Room = 'root';
    obj.Event = 'add';
    sendMsg(obj)
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
    a.src = "javascript:abrirSala('"+'Nome da Sala'+"')";
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
        };
        conn.onmessage = function (evt) {
            msgEvt = JSON.parse(evt.data)
            if(msgEvt.event == "msg"){
                appendChat(msgEvt);
                addMsg(msgEvt);
            } else {
                // tratar outros tipos de mensagens
                console.log(msgEvt)
            }
        };
    } else {
        console.log("Your browser does not support WebSockets.")
    }
}