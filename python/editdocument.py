from flask import Flask, request, jsonify, send_file
import os
from threading import Thread
from docx import Document

app = Flask(__name__)

# Путь для сохранения временных файлов
UPLOAD_FOLDER = 'uploads'

# Редактирование файла
def edit_docx(filepath):
    # Открываем документ
    doc = Document(filepath)
    # Добавляем текст в конец документа
    doc.add_paragraph('Это я добавил текст, хы хы!.')
    # Сохраняем изменения
    edited_filepath = filepath.replace('.docx', '_edited.docx')
    doc.save(edited_filepath)
    return edited_filepath

def cleanup_temp_files():
    # Перебираем все файлы в папке загрузок
    if os.path.exists(UPLOAD_FOLDER):
        for filename in os.listdir(UPLOAD_FOLDER):
            file_path = os.path.join(UPLOAD_FOLDER, filename)
            try:
                # Удаляем файл
                os.remove(file_path)
                print(f"Deleted unused file: {file_path}")
            except OSError:
                # Если файл занят, игнорируем его
                print(f"File is in use, skipping: {file_path}")

        # Проверяем, остались ли файлы в папке
        if not os.listdir(UPLOAD_FOLDER):
            # Если папка пуста, удаляем её
            os.rmdir(UPLOAD_FOLDER)
            print(f"Deleted empty directory: {UPLOAD_FOLDER}")

@app.route('/editdocx', methods=['POST'])
def edit_docx_route():
    # Проверяем, что файл был отправлен
    if 'file' not in request.files:
        return jsonify({'error': 'No file provided'}), 400

    file = request.files['file']
    if file.filename == '':
        return jsonify({'error': 'No selected file'}), 400

    # Создаем папку uploads, если её нет
    if not os.path.exists(UPLOAD_FOLDER):
        os.makedirs(UPLOAD_FOLDER)

    # Сохраняем загруженный файл
    filepath = os.path.join(UPLOAD_FOLDER, file.filename)
    try:
        file.save(filepath)
        print(f"File saved successfully: {filepath}")
    except Exception as e:
        return jsonify({'error': f'Failed to save file: {str(e)}'}), 500

    # Редактируем документ
    edited_filepath = edit_docx(filepath)

    # Отправляем измененный файл обратно
    response = send_file(edited_filepath, as_attachment=True)

    # Запускаем поток очистки после отправки файла
    Thread(target=cleanup_temp_files, daemon=True).start()

    return response

if __name__ == '__main__':
    app.run(debug=True, port=5000)
