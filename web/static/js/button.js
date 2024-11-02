// после первого нажатия кнопка становится неактивной
function handleClick() {
    const button = document.getElementById('edit-button-download');
    button.style.pointerEvents = 'none';
    button.style.backgroundColor = 'rgb(99, 185, 242)';
}