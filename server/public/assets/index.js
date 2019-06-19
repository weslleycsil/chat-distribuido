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
var nomeSala = document.getElementById("namesala");
var popUsers = document.getElementById("listUsers");
var popRooms = document.getElementById("listRooms");

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
    obj.Room = 'root';
    sendMsg(obj);
    
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
        getListRooms();
        $('#joinRoom').modal('show')
    } else if (n ==3){
        $('#changeNick').modal('show')
    } else if (n ==4){
        getListUsers();
        $('#ListUsersRoom').modal('show')
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
 * Funcionamento OK
 */
function leaveRoom(){
    var ObjChat = JSON.parse(localStorage.getItem('ObjChat'));
    if(ObjChat.RoomActive == 'root'){
        alert('Você não pode sair da Sala Root');
        return false
    }
    var Rooms = JSON.parse(localStorage.getItem('Rooms'));
    const i = Rooms.indexOf(ObjChat.RoomActive)
    N = Rooms.splice(i,1);
    localStorage.setItem('Rooms', JSON.stringify(Rooms));
    obj.Event = 'leave';
    obj.Room = ObjChat.RoomActive;
    obj.Message = "sair da sala";
    var item = document.getElementById(ObjChat.RoomActive);
    roomList.removeChild(item);
    sendMsg(obj);
    // mudar para a sala root
    abrirSala("root");

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
    document.getElementById("inputNewRoom").value = "";
    $('#NewRoom').modal('hide');
};

/**
 * Função para entrar em um sala que eu ainda não entrei
 * @param {string} nome 
 * Funcionamento OK
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
    appendRoom(nome);
    $('#joinRoom').modal('hide');
};

/**
 * Função para mudar o Username
 * Funcionamento OK
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

/**
 * Função para adiconar uma sala na lista lateral
 * @param {string} room 
 * Funcionamento OK
 */
function appendRoom(room){
    var item = document.createElement("li");
    item.id = room;
    var a = document.createElement("a");
    a.href = "javascript:abrirSala('"+room+"')";
    a.innerText = '#'+room;
    item.appendChild(a);
    roomList.appendChild(item);
};

/**
 *  Funcão para adiconar a sala no local storage e tbm acionar a colocação na barra lateral
 * @param {string} sala
 * Funcionamento OK
 */
function addSala(nome){
    var Rooms = JSON.parse(localStorage.getItem('Rooms'));
    const i = Rooms.indexOf(nome);
    if(i != -1){
        //sala já existe
        return false
    }
    //posso criar a sala
    Rooms.push(nome);
    appendRoom(nome);
    localStorage.setItem("Rooms", JSON.stringify(Rooms));
    return true;
}

/**
 * Funcão para abrir a sala no frontend
 * @param {string} sala 
 * Funcionamento OK
 */
function abrirSala(sala){
    var ObjChat = JSON.parse(localStorage.getItem('ObjChat'));
    if(ObjChat.RoomActive == sala){
        return false;
    }
    obj.Room = sala;
    ObjChat.RoomActive = sala;
    localStorage.setItem('ObjChat', JSON.stringify(ObjChat));

    //apagas as msg da tela
    c = chat.children;
    while(c.length > 1){
        chat.removeChild(c[1]);
    }
    //mudar nome da sala titulo
    nomeSala.innerText = 'Sala #'+sala;
    //adicionar msg das salas na tela
    var msgs = JSON.parse(localStorage.getItem(sala));
    if(msgs == null){
        msgs = [];
    }
    msgs.forEach(element => {
        appendChat(element);
    });
}

/**
 * Função para adicionar mensagens no local storage de cada sala
 * @param {*} msg 
 * Funcionamento OK
 */
function addMsg(msg){
    var msgs = JSON.parse(localStorage.getItem(msg.room));
    if(msgs == null){
        msgs = [];
    }
    msgs.push(msg);
    localStorage.setItem(msg.room, JSON.stringify(msgs));
};

/**
 * Função para enviar uma nova mensagem
 * Funcionamento OK
 */
function Enviar(){
    if (!conn) {
        return false;
    }
    var ObjChat = JSON.parse(localStorage.getItem('ObjChat'));
    obj.Room = ObjChat.RoomActive;
    obj.Message = document.getElementById("msg").value;
    obj.Event = "msg";
    sendMsg(obj);
    document.getElementById("msg").value = "";
    return true;
}

function getListRooms(){
    obj.Event = 'listRooms';
    obj.Message = "listar salas";
    obj.Room = "";
    sendMsg(obj);
}

function getListUsers(){
    var ObjChat = JSON.parse(localStorage.getItem('ObjChat'));
    obj.Event = 'listUsers';
    obj.Message = "listar usuarios";
    obj.Room = ObjChat.RoomActive;
    sendMsg(obj);
}

function listUsers(lista){
    c = popUsers.children;
    while(c.length > 0){
        popUsers.removeChild(c[0]);
    }
    var l = lista.split(",");
    var newList = [];
    l.forEach(element=>{
        if(element != ""){
            var item = document.createElement("li");
            item.innerText = element;
            popUsers.appendChild(item);
        }
    });
}

function listRooms(lista){
    c = popRooms.children;
    while(c.length > 0){
        popRooms.removeChild(c[0]);
    }
    var l = lista.split(",");
    var newList = [];
    l.forEach(element=>{
        if(element != ""){
            var item = document.createElement("li");
            var a = document.createElement("a");
            a.href = "javascript:joinRoom('"+element+"')";
            a.innerText = 'Sala #'+element;
            item.appendChild(a);
            popRooms.appendChild(item);
        }
    });
}


window.onload = function () {
    $('#enter').modal('show')
    var ObjChat = {
        Username: null,
        Email: null,
        RoomActive: null,
    };
    localStorage.setItem('ObjChat', JSON.stringify(ObjChat));
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
                addMsg(msgEvt);
                var ObjChat = JSON.parse(localStorage.getItem('ObjChat'));
                if(ObjChat.RoomActive == null){
                    ObjChat.RoomActive= "root";
                }
                if(msgEvt.room == ObjChat.RoomActive){
                    appendChat(msgEvt); // ver melhor recepção
                }
            } else if(msgEvt.event == "command" && msgEvt.message == "add sala"){
                //adicionar sala ao array de salas
                addSala(msgEvt.room)
            } else if(msgEvt.event == "listRooms"){
                console.log("Listar Rooms: ", msgEvt.message);
                listRooms(msgEvt.message)
            } else if(msgEvt.event == "listUsers"){
                console.log("Listar Users: ", msgEvt.message);
                listUsers(msgEvt.message)
            } else {
                // tratar outros tipos de mensagens
                console.log(msgEvt)
            }
        };
    } else {
        console.log("Your browser does not support WebSockets.")
    }
}