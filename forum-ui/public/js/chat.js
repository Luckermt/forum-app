class ChatService {
    static socket = null;
    static token = null;
    
    static init(token) {
        this.token = token;
        this.connect();
        this.setupEventListeners();
    }
    
    static connect() {
        if (this.socket) {
            this.socket.close();
        }
        
        this.socket = new WebSocket(`ws://${window.location.host}/ws?token=${this.token}`);
        
        this.socket.onopen = () => {
            console.log('WebSocket connected');
            document.getElementById('message-input').disabled = false;
            document.getElementById('send-btn').disabled = false;
        };
        
        this.socket.onclose = () => {
            console.log('WebSocket disconnected');
            document.getElementById('message-input').disabled = true;
            document.getElementById('send-btn').disabled = true;
            
            // Попытка переподключения через 5 секунд
            setTimeout(() => this.connect(), 5000);
        };
        
        this.socket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.displayMessage(message);
        };
        
        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    static setupEventListeners() {
        // Отправка сообщения по нажатию кнопки
        document.getElementById('send-btn').addEventListener('click', () => {
            this.sendMessage();
        });
        
        // Отправка сообщения по Enter
        document.getElementById('message-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.sendMessage();
            }
        });
    }
    
    static sendMessage() {
        const input = document.getElementById('message-input');
        const text = input.value.trim();
        
        if (text && this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify({ text }));
            input.value = '';
        }
    }
    
    static displayMessage(message) {
        const messagesContainer = document.getElementById('chat-messages');
        const messageElement = document.createElement('div');
        messageElement.className = 'message';
        
        messageElement.innerHTML = `
            <div class="author">${message.username}</div>
            <div class="content">${message.text}</div>
            <div class="time">${new Date(message.timestamp).toLocaleTimeString()}</div>
        `;
        
        messagesContainer.appendChild(messageElement);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
    
    static updateOnlineCount(count) {
        document.getElementById('online-count').textContent = count;
    }
}