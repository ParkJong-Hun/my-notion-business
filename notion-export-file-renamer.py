import os
import re
import zipfile
import shutil

file_regexp = r'( \w{32,32}\b)'  # 파일 이름에서 사용한 정규식 패턴입니다.

def process_markdown_files():
    for root, dirs, files in os.walk(os.getcwd(), topdown=False):
        for f in files:
            f_path = os.path.join(root, f)
            if f.endswith('.md'):  # .md 파일인지 확인합니다.
                with open(f_path, 'r', encoding='utf-8') as file:
                    content = file.read()
                substituted_content = re.sub(r'\[.*?\]\((.*?)\)', lambda x: '[' + os.path.splitext(os.path.basename(x.group(1)))[0] + '](' + os.path.splitext(os.path.basename(x.group(1)))[0] + ')', content)
                # 위의 코드는 []() 패턴을 찾아서 파일명과 동일한 조건으로 변경합니다.
                with open(f_path, 'w', encoding='utf-8') as file:
                    file.write(substituted_content)

                # 파일 내의 링크 URL에 적용한 후에 파일명 변경을 수행합니다.
                substitute_f_path = re.sub(file_regexp, '', os.path.abspath(f_path))
                os.renames(f_path, substitute_f_path)

def extract_and_replace_src():
    root_dir = os.getcwd()
    raw_folder = os.path.join(root_dir, 'raw')

    # raw 폴더에서 zip 파일 검색
    zip_files = [f for f in os.listdir(raw_folder) if f.endswith('.zip')]

    if not zip_files:
        print("No zip files found in the 'raw' folder.")
        return

    # 첫 번째 zip 파일 선택
    zip_file_path = os.path.join(raw_folder, zip_files[0])

    # 압축 해제
    extract_path = os.path.join(root_dir, 'temp_extract')
    with zipfile.ZipFile(zip_file_path, 'r') as zip_ref:
        zip_ref.extractall(extract_path)

    # Notion이 내보내기 zip을 또 다른 zip으로 감싸는 새 형식 처리
    extracted_items = os.listdir(extract_path)
    if len(extracted_items) == 1 and extracted_items[0].endswith('.zip'):
        inner_zip_path = os.path.join(extract_path, extracted_items[0])
        inner_extract_path = os.path.join(root_dir, 'temp_extract_inner')
        with zipfile.ZipFile(inner_zip_path, 'r') as zip_ref:
            zip_ref.extractall(inner_extract_path)
        shutil.rmtree(extract_path)
        extract_path = inner_extract_path

    # 기존 src 폴더 삭제
    src_path = os.path.join(root_dir, 'src')
    if os.path.exists(src_path):
        shutil.rmtree(src_path)

    # 압축 해제된 전체 폴더를 src로 이동
    os.rename(extract_path, src_path)

if __name__ == "__main__":
    extract_and_replace_src()
    process_markdown_files()
