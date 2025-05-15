// Toggle mobile sidebar
document.getElementById('mobile-menu-button').addEventListener('click', function() {
    document.querySelector('.sidebar').classList.toggle('open');
});

// Active link highlighting
const currentPath = window.location.pathname;
document.querySelectorAll('.sidebar-link').forEach(link => {
    if (link.getAttribute('href') === currentPath) {
        link.classList.add('active');
    }
});

// Product image preview
document.querySelectorAll('.product-image-input').forEach(input => {
    input.addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(event) {
                document.getElementById('product-image-preview').src = event.target.result;
            };
            reader.readAsDataURL(file);
        }
    });
});