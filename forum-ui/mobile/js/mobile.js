class MobileUI {
    static init() {
        if (!this.isMobile()) return;
        
        this.addMobileMenuButton();
        this.adaptUIForMobile();
    }
    
    static isMobile() {
        return /Mobi|Android/i.test(navigator.userAgent);
    }
    
    static addMobileMenuButton() {
        const menuButton = document.createElement('button');
        menuButton.id = 'mobile-menu-btn';
        menuButton.innerHTML = '<i class="fas fa-bars"></i>';
        menuButton.className = 'btn btn-small';
        
        const header = document.querySelector('.header');
        if (header) {
            header.insertBefore(menuButton, header.querySelector('#user-info'));
        }
        
        menuButton.addEventListener('click', () => {
            document.querySelector('.sidebar').classList.toggle('mobile-visible');
        });
    }
    
    static adaptUIForMobile() {
        // Добавляем класс для мобильных стилей
        document.body.classList.add('mobile');
        
        // Обработчик для закрытия меню при клике на контент
        document.querySelector('.content').addEventListener('click', () => {
            document.querySelector('.sidebar').classList.remove('mobile-visible');
            document.getElementById('admin-panel').classList.remove('mobile-visible');
        });
        
        // Кнопка для открытия админ-панели (если пользователь админ)
        const user = AuthService.getCurrentUser();
        if (user?.role === 'admin') {
            const adminButton = document.createElement('button');
            adminButton.id = 'mobile-admin-btn';
            adminButton.innerHTML = '<i class="fas fa-cog"></i>';
            adminButton.className = 'btn btn-small';
            
            const header = document.querySelector('.header');
            if (header) {
                header.insertBefore(adminButton, header.querySelector('#user-info'));
            }
            
            adminButton.addEventListener('click', () => {
                document.getElementById('admin-panel').classList.toggle('mobile-visible');
            });
        }
    }
}

// Инициализация мобильного UI
document.addEventListener('DOMContentLoaded', () => {
    MobileUI.init();
});