class AdminService {
    static token = null;
    
    static init(token) {
        this.token = token;
        this.setupEventListeners();
        this.loadUsers();
    }
    
    static setupEventListeners() {
        // Переключение вкладок
        document.querySelectorAll('.admin-tabs .tab-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.admin-tabs .tab-btn').forEach(b => b.classList.remove('active'));
                document.querySelectorAll('.admin-tab-content').forEach(c => c.classList.add('hidden'));
                
                btn.classList.add('active');
                const tab = btn.getAttribute('data-tab');
                document.getElementById(`${tab}-tab`).classList.remove('hidden');
            });
        });
        
        // Закрытие админ-панели
        document.getElementById('close-admin-btn').addEventListener('click', () => {
            document.getElementById('admin-panel').classList.add('hidden');
        });
        
        // Поиск пользователей
        document.getElementById('user-search-btn').addEventListener('click', () => {
            this.loadUsers();
        });
    }
    
    static async loadUsers() {
        try {
            const searchQuery = document.getElementById('user-search').value;
            const response = await fetch(`/admin/users?search=${encodeURIComponent(searchQuery)}`, {
                headers: { 'Authorization': `Bearer ${this.token}` }
            });
            
            if (!response.ok) throw new Error('Ошибка загрузки пользователей');
            
            const users = await response.json();
            this.renderUsers(users);
        } catch (error) {
            console.error('Error loading users:', error);
            showNotification('Ошибка загрузки пользователей', 'error');
        }
    }
    
    static renderUsers(users) {
        const usersList = document.getElementById('users-list');
        usersList.innerHTML = '';
        
        users.forEach(user => {
            const row = document.createElement('tr');
            
            row.innerHTML = `
                <td>${user.username}</td>
                <td>${user.email}</td>
                <td>${user.blocked ? 'Заблокирован' : 'Активен'}</td>
                <td>
                    <button class="btn btn-small ${user.blocked ? 'btn-primary' : 'btn-danger'} block-btn" data-user-id="${user.id}">
                        ${user.blocked ? 'Разблокировать' : 'Заблокировать'}
                    </button>
                </td>
            `;
            
            usersList.appendChild(row);
        });
        
        // Обработчики кнопок блокировки
        document.querySelectorAll('.block-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const userId = btn.dataset.userId;
                const action = btn.textContent === 'Заблокировать' ? 'block' : 'unblock';
                this.toggleUserBlock(userId, action);
            });
        });
    }
    
    static async toggleUserBlock(userId, action) {
        try {
            const response = await fetch(`/admin/users/${userId}/${action}`, {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${this.token}` }
            });
            
            if (!response.ok) throw new Error(`Ошибка ${action === 'block' ? 'блокировки' : 'разблокировки'} пользователя`);
            
            this.loadUsers();
            showNotification(`Пользователь успешно ${action === 'block' ? 'заблокирован' : 'разблокирован'}`, 'success');
        } catch (error) {
            console.error(`Error ${action === 'block' ? 'blocking' : 'unblocking'} user:`, error);
            showNotification(`Ошибка ${action === 'block' ? 'блокировки' : 'разблокировки'} пользователя`, 'error');
        }
    }
}