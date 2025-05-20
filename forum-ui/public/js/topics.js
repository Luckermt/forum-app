class TopicsService {
    static currentPage = 1;
    static itemsPerPage = 10;
    static token = null;
    
    static init(token) {
        this.token = token;
        this.setupEventListeners();
        this.loadTopics();
    }
    
    static setupEventListeners() {
        // Новая тема
        document.getElementById('new-topic-btn').addEventListener('click', () => {
            this.showNewTopicModal();
        });
        
        // Пагинация
        document.getElementById('prev-page').addEventListener('click', () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.loadTopics();
            }
        });
        
        document.getElementById('next-page').addEventListener('click', () => {
            this.currentPage++;
            this.loadTopics();
        });
        
        // Поиск тем
        document.getElementById('search-btn').addEventListener('click', () => {
            this.currentPage = 1;
            this.loadTopics();
        });
        
        // Создание темы
        document.getElementById('create-topic-btn').addEventListener('click', () => {
            this.createTopic();
        });
        
        // Назад к чату
        document.getElementById('back-to-chat').addEventListener('click', () => {
            document.getElementById('chat-container').classList.remove('hidden');
            document.getElementById('topic-container').classList.add('hidden');
        });
    }
    
    static async loadTopics() {
        try {
            const searchQuery = document.getElementById('topic-search').value;
            const response = await fetch(`/topics?page=${this.currentPage}&limit=${this.itemsPerPage}&search=${encodeURIComponent(searchQuery)}`, {
                headers: { 'Authorization': `Bearer ${this.token}` }
            });
            
            if (!response.ok) throw new Error('Ошибка загрузки тем');
            
            const { topics, total } = await response.json();
            this.renderTopics(topics, total);
        } catch (error) {
            console.error('Error loading topics:', error);
            showNotification('Ошибка загрузки тем', 'error');
        }
    }
    
    static renderTopics(topics, total) {
        const topicsList = document.getElementById('topics-list');
        topicsList.innerHTML = '';
        
        topics.forEach(topic => {
            const topicElement = document.createElement('div');
            topicElement.className = 'topic-item';
            topicElement.dataset.id = topic.id;
            
            topicElement.innerHTML = `
                <h3>${topic.title}</h3>
                <div class="meta">
                    <span>Автор: ${topic.username}</span>
                    <span>Сообщений: ${topic.messageCount}</span>
                </div>
            `;
            
            topicElement.addEventListener('click', () => {
                this.loadTopicMessages(topic.id, topic.title);
            });
            
            topicsList.appendChild(topicElement);
        });
        
        // Обновление пагинации
        document.getElementById('page-info').textContent = `Страница ${this.currentPage}`;
        document.getElementById('prev-page').disabled = this.currentPage <= 1;
        document.getElementById('next-page').disabled = this.currentPage * this.itemsPerPage >= total;
    }
    
    static async loadTopicMessages(topicId, topicTitle) {
        try {
            const response = await fetch(`/messages?topic_id=${topicId}`, {
                headers: { 'Authorization': `Bearer ${this.token}` }
            });
            
            if (!response.ok) throw new Error('Ошибка загрузки сообщений');
            
            const messages = await response.json();
            this.renderTopicMessages(topicId, topicTitle, messages);
        } catch (error) {
            console.error('Error loading topic messages:', error);
            showNotification('Ошибка загрузки сообщений темы', 'error');
        }
    }
    
    static renderTopicMessages(topicId, topicTitle, messages) {
        document.getElementById('chat-container').classList.add('hidden');
        document.getElementById('topic-container').classList.remove('hidden');
        
        document.getElementById('topic-title').textContent = topicTitle;
        document.getElementById('topic-message-input').dataset.topicId = topicId;
        
        const messagesContainer = document.getElementById('topic-messages');
        messagesContainer.innerHTML = '';
        
        messages.forEach(message => {
            const messageElement = document.createElement('div');
            messageElement.className = 'message';
            
            messageElement.innerHTML = `
                <div class="author">${message.username}</div>
                <div class="content">${message.content}</div>
                <div class="time">${new Date(message.createdAt).toLocaleString()}</div>
            `;
            
            messagesContainer.appendChild(messageElement);
        });
        
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
        
        // Показать кнопку удаления для админов
        const user = AuthService.getCurrentUser();
        document.getElementById('delete-topic-btn').classList.toggle('hidden', user?.role !== 'admin');
        document.getElementById('delete-topic-btn').dataset.topicId = topicId;
        document.getElementById('delete-topic-btn').addEventListener('click', () => {
            this.deleteTopic(topicId);
        });
        
        // Обработчик отправки сообщения в тему
        document.getElementById('send-topic-message-btn').addEventListener('click', () => {
            this.sendTopicMessage(topicId);
        });
        
        document.getElementById('topic-message-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.sendTopicMessage(topicId);
            }
        });
    }
    
    static showNewTopicModal() {
        document.getElementById('new-topic-modal').classList.remove('hidden');
        document.getElementById('topic-title-input').focus();
    }
    
    static hideNewTopicModal() {
        document.getElementById('new-topic-modal').classList.add('hidden');
        document.getElementById('topic-title-input').value = '';
        document.getElementById('topic-content-input').value = '';
    }
    
    static async createTopic() {
        const title = document.getElementById('topic-title-input').value.trim();
        const content = document.getElementById('topic-content-input').value.trim();
        
        if (!title || !content) {
            showNotification('Заполните все поля', 'error');
            return;
        }
        
        try {
            const response = await fetch('/topics', {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                body: JSON.stringify({ title, content })
            });
            
            if (!response.ok) throw new Error('Ошибка создания темы');
            
            const topic = await response.json();
            this.hideNewTopicModal();
            this.loadTopics();
            showNotification('Тема успешно создана', 'success');
        } catch (error) {
            console.error('Error creating topic:', error);
            showNotification('Ошибка создания темы', 'error');
        }
    }
    
    static async deleteTopic(topicId) {
        if (!confirm('Вы уверены, что хотите удалить эту тему?')) return;
        
        try {
            const response = await fetch(`/topics/${topicId}`, {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${this.token}` }
            });
            
            if (!response.ok) throw new Error('Ошибка удаления темы');
            
            document.getElementById('chat-container').classList.remove('hidden');
            document.getElementById('topic-container').classList.add('hidden');
            this.loadTopics();
            showNotification('Тема удалена', 'success');
        } catch (error) {
            console.error('Error deleting topic:', error);
            showNotification('Ошибка удаления темы', 'error');
        }
    }
    
    static async sendTopicMessage(topicId) {
        const input = document.getElementById('topic-message-input');
        const content = input.value.trim();
        
        if (!content) return;
        
        try {
            const response = await fetch('/messages', {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                body: JSON.stringify({ 
                    topicId,
                    content,
                    isChat: false 
                })
            });
            
            if (!response.ok) throw new Error('Ошибка отправки сообщения');
            
            input.value = '';
            this.loadTopicMessages(topicId, document.getElementById('topic-title').textContent);
        } catch (error) {
            console.error('Error sending message:', error);
            showNotification('Ошибка отправки сообщения', 'error');
        }
    }
}

// Инициализация модального окна
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('close-modal')?.addEventListener('click', () => {
        TopicsService.hideNewTopicModal();
    });
    
    document.getElementById('cancel-topic-btn')?.addEventListener('click', () => {
        TopicsService.hideNewTopicModal();
    });
});