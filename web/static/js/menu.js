const buttons = document.querySelectorAll(".window-one-button");

buttons.forEach(button => {
  button.addEventListener("click", function (event) {
    event.preventDefault(); // Отключает переход по ссылке
    
    // Убираем класс active у всех кнопок
    buttons.forEach(b => b.classList.remove("active"));
    
    // Добавляем класс active на текущую кнопку
    button.classList.add("active");
  });
});