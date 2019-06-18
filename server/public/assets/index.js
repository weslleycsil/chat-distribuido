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

/**
 * Local Storage Gabarito
 * Objetos: ObjChat, rooms, nomeDaSala
 * 
 * ObjChat -> {"Username": "nome de usuario", "Email": "email do usuario", "RoomActive": "sala ativa"}
 * nomeDaSala -> [{email: "email", username: "Servidor", message: "Cleiton Entrou.", event: "msg", room: "root"}]
 * Rooms -> ["root","room2"]
 * 
 */

/**
 * Verificar Funcionamento
 * 
 * Função com o intuito de iniciar todo o chat
 * envia o pedido para mudar o username
 * envia o pedido para criar a sala root
 */
function enterChat() {
    var ObjChat = {
        Username: document.getElementById("inputUser").value,
        Email: document.getElementById("inputEmail").value,
        RoomActive: 'root',
    };
    var Rooms = ['root'];

    localStorage.setItem('ObjChat', JSON.stringify(ObjChat));
    localStorage.setItem('Rooms', JSON.stringify(Rooms));
    obj.Username = ObjChat.Username;
    obj.Message = "mudar nome";
    obj.Email = ObjChat.Email;
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

/**
 * Função para abrir uma determinada PopUp
 * Funcionamento OK
 * 
 * @param {int} n numero da popup a ser aberta
 */
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
};

/**
 * Função genérica para envio de mensagens para o websocket
 * Funcionamento OK
 * 
 * @param {*} msg 
 */
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

/**
 * Função para retornar a url da imagem do avatar
 * Funcionamento OK
 * 
 * @param {string} email 
 */
function gravatar(email) {
    return 'http://www.gravatar.com/avatar/' + md5(email);
};

/**
 * Função para deixar uma determinada Sala Ativa
 * Você não sai da Sala Root
 * Funcionamento (falta pular para a sala root)
 */
function leaveRoom(){
    var ObjChat = JSON.parse(localStorage.getItem('ObjChat'));
    if(ObjChat.RoomActive == 'root'){
        alert('Você não pode sair da Sala Root');
        return false
    }
    var Rooms = JSON.parse(localStorage.getItem('Rooms'));
    const i = Rooms.indexOf(ObjChat.RoomActive)
    NewRooms = Rooms.splice(i,1);
    localStorage.setItem('Rooms', JSON.stringify(NewRooms));
    obj.Event = 'leave';
    obj.Room = ObjChat.RoomActive;
    obj.Message = "sair da sala";
    var item = document.getElementById(ObjChat.RoomActive);
    roomList.removeChild(item);
    sendMsg(obj);
    // mudar para a sala root

}

/**
 * Função para Criar uma nova Sala
 * Sempre que eu crio uma sala, eu também entro nela
 * Funcionamento OK
 */
function newRoom() {
    //criar uma nova sala
    var roomNew = document.getElementById("inputNewRoom").value;
    obj.Event = 'add';
    obj.Message = "adicionar sala";
    obj.Room = roomNew;
    sendMsg(obj);
    $('#NewRoom').modal('hide');
};

/**
 * Função para entrar em um sala que eu ainda não entrei
 * @param {string} nome 
 */
function joinRoom(nome) {
    //verifico se já estou na sala
    var Rooms = JSON.parse(localStorage.getItem('Rooms'));
    const i = Rooms.indexOf(nome);
    if(i != -1){
        alert("Você já está nessa sala!")
        return false;
    }
    //se não, eu entro na sala
    Rooms.push(nome);
    localStorage.setItem('Rooms', JSON.stringify(Rooms));
    //enviar msg para join new room
    obj.Event = 'join';
    obj.Room = nome;
    obj.Message = "entrar em sala";
    sendMsg(obj);
    $('#joinRoom').modal('hide');
};

/**
 * Função para mudar o Username
 */
function changeUsername() {
    var ObjChat = JSON.parse(localStorage.getItem('ObjChat'));
    ObjChat.Username = document.getElementById("inputUsername").value;
    localStorage.setItem('ObjChat', JSON.stringify(ObjChat));
    //enviar msg para join new room
    obj.Event = 'change';
    obj.Room = ObjChat.RoomActive;
    obj.Username = chat.Username;
    obj.Message = "mudar nome";
    sendMsg(obj);
    $('#changeNick').modal('hide');
};

/**
 * Função para adicionar a função sair da sala ao rodape do chat
 * Funcionamento OK
 */
function appendSair(){
    var small = document.createElement("small");
    small.className = "d-block text-right mt-3";
    var a = document.createElement("a");
    a.href = "javascript:leaveRoom()";
    a.innerText = 'Sair da Sala';
    small.appendChild(a);
    chat.appendChild(small);
}



/**
 * Função para adicionar conversas no chat
 * @param {*} m 
 * Funcionamento OK
 */
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
};



function appendRoom(room){
    var item = document.createElement("li");
    item.id = room;
    var a = document.createElement("a");
    a.href = "javascript:abrirSala('"+room+"')";
    a.innerText = 'Sala #'+room;
    item.appendChild(a);
    roomList.appendChild(item);
}; // verificar se já existe a sala antes de adicionar na lista



function addMsg(msg){
    var msgs = JSON.parse(localStorage.getItem(msg.room));
    if(msgs == null){
        msgs = [];
    }
    msgs.push(msg);
    localStorage.setItem(msg.room, JSON.stringify(msgs));
};

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
    //verificar se já esta na sala ou nao
    //se nao, join room

    //apagas as msg da tela
    c = chat.children;
    while(c.length > 2 && c[1].nodeName == "DIV"){
        chat.removeChild(c[1]);
    }
    //adicionar msg das salas na tela
    var msgs = JSON.parse(localStorage.getItem(sala));
    if(msgs == null){
        msgs = [];
    }
    msgs.forEach(element => {
        appendChat(element);
    });
}

function Enviar(){
    if (!conn) {
        return false;
    }
    var chat = JSON.parse(localStorage.getItem('objChat'));
    obj.Room = chat.room;
    obj.Message = document.getElementById("msg").value;
    obj.Event = "msg";
    sendMsg(obj);
    document.getElementById("msg").value = "";
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
            console.log("Nova MSG recebida: ", msgEvt)
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