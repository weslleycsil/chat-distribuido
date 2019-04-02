var obj = {
    Email: '',
    Username: '',
    Message: '',
    Event: '',
    Room: ''
};

var conn; // conexao websocket


function newRoom() {
    //criar uma nova sala
    var roomNew = document.getElementById("").value;
    obj.Event = 'add';
    obj.Room = roomNew;
    sendMsg(obj);
    console.log('New Room!');
};

function joinRoom() {
    var chat = localStorage.getItem('objChat')
    chat.Room = document.getElementById("").value;
    localStorage.setItem('objChat', chat);
    //enviar msg para join new room
    obj.Event = 'join';
    obj.Room = chat.Room;
    sendMsg(obj);
    console.log('Join Room!');
};

function enterChat() {
    var chat = {
        email: document.getElementById("").value,
        username: document.getElementById("").value,
        room: '',
    };
    localStorage.setItem('objChat', chat);
    obj.Username = chat.username;
    obj.Email = chat.email;
    obj.Event = 'change';
    sendMsg(obj);
    console.log('Enter Chat!');
};

function changeUsername() {
    var chat = localStorage.getItem('objChat')
    chat.Username = document.getElementById("").value;
    localStorage.setItem('objChat', chat);
    //enviar msg para join new room
    obj.Event = 'change';
    sendMsg(obj);
    console.log('Change Username!');
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