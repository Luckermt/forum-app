class AuthService {
    static async login(email, password) {
        try {
            const response = await fetch('/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Ошибка авторизации');
            }
            
            const data = await response.json();
            localStorage.setItem('forum_token', data.token);
            localStorage.setItem('forum_user', JSON.stringify(data.user));
            
            return data;
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    }
    
    static async register(username, email, password) {
    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({ username, email, password })
        });
        
        if (!response.ok) {
            // Попробуем получить текст ошибки
            const errorText = await response.text();
            try {
                const errorJson = JSON.parse(errorText);
                throw new Error(errorJson.message || 'Registration failed');
            } catch {
                throw new Error(errorText || 'Registration failed');
            }
        }
        
        // Очистка ответа перед парсингом
        const responseText = await response.text();
        const cleanResponse = responseText.trim();
        return JSON.parse(cleanResponse);
    } catch (error) {
        console.error('Register error:', error);
        throw error;
    }
}
    
    static logout() {
        localStorage.removeItem('forum_token');
        localStorage.removeItem('forum_user');
        window.location.reload();
    }
    
    static getCurrentUser() {
        const user = localStorage.getItem('forum_user');
        return user ? JSON.parse(user) : null;
    }
}

// Инициализация обработчиков авторизации
document.addEventListener('DOMContentLoaded', () => {
    // Обработчик входа
    document.getElementById('login-btn')?.addEventListener('click', async () => {
        const email = document.getElementById('login-email').value;
        const password = document.getElementById('login-password').value;
        const errorElement = document.getElementById('login-error');
        
        try {
            errorElement.textContent = '';
            await AuthService.login(email, password);
            window.location.reload();
        } catch (error) {
            errorElement.textContent = error.message;
        }
    });
    
    // Обработчик регистрации
    document.getElementById('register-btn')?.addEventListener('click', async () => {
        const username = document.getElementById('reg-username').value;
        const email = document.getElementById('reg-email').value;
        const password = document.getElementById('reg-password').value;
        const confirm = document.getElementById('reg-confirm').value;
        const errorElement = document.getElementById('register-error');
        
        try {
            errorElement.textContent = '';
            
            if (password !== confirm) {
                throw new Error('Пароли не совпадают');
            }
            
            await AuthService.register(username, email, password);
            console.log('Registration success:', data);
            showNotification('Регистрация успешна! Теперь вы можете войти.', 'success');
            
            

            // Переключиться на вкладку входа
            document.querySelector('.auth-tabs .tab-btn[data-tab="login"]').click();
            document.getElementById('login-email').value = email;
            document.getElementById('login-password').value = '';
            document.getElementById('reg-username').value = '';
            document.getElementById('reg-email').value = '';
            document.getElementById('reg-password').value = '';
            document.getElementById('reg-confirm').value = '';
        } catch (error) {
            console.error('Registration error:', error);
            errorElement.textContent = error.message;
            errorElement.style.color = 'red';
        }
    });
    
    // Обработчик выхода
    document.getElementById('logout-btn')?.addEventListener('click', () => {
        AuthService.logout();
    });
});