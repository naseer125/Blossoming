# AI Rules - Blossoming

## 스크립트 처리 규칙

### Python 스크립트 (`widen_gracefully.py`)

**우선순위**: Python 버전을 먼저 고려

**실행 방법**:
```bash
# 단일 이미지
python3 widen_gracefully.py 입력파일.jpg

# 폴더 전체 (권장)
python3 widen_gracefully.py /path/to/folder
python3 widen_gracefully.py .  # 현재 폴더
```

**중요 규칙**:
1. uv 가상환경 사용: `/Users/crong/Pictures/Wallpaper/.venv`
2. 의존성: Pillow, numpy, opencv-python
3. Python 버전: 3.13.11
4. 결과물 저장 경로: `python/` 폴더

**이미지 방향별 처리 규칙**:
1. **세로형 (Portrait)**:
   - 워터마크 제거
   - 여백 제거
   - 리사이즈 (높이 2160)
   - 16:9 변환 (좌우 블러 확장)

2. **가로형 (Landscape)**:
   - 스마트 크롭 (3단계 폴백 체인)
   - 리사이즈 (너비 3840)

3. **정방형 (Square)**:
   - 자동 건너뜀

**스마트 크롭 규칙 (3단계 폴백)**:
1. **Step 1: 얼굴 감지**
   - OpenCV Haar Cascade 사용
   - 배경색 기반 머리카락 끝 감지 (다중 스캔 라인)
   - 얼굴이 상단 1/3 지점에 오도록 크롭
   - 실패 시 Step 2로 이동

2. **Step 2: 몸통 감지**
   - HOG Detector 사용
   - 가장 상단 몸통 박스 선택
   - 몸통 상단에서 20% 만큼 위로 머리 추정
   - 실패 시 Step 3로 이동

3. **Step 3: 중앙 크롭 (폴백)**
   - 이미지 중앙을 기준으로 16:9 비율 크롭
   - 안전한 대안으로 항상 성공

**코드 수정 시 확인사항**:
- `ImageConverter` 클래스 내에서만 수정
- 메모리 처리 방식 유지 (임시 파일 사용 금지)
- ICC 프로필 보존 로직 필수 유지
- `process_image()` 메서드 순서 변경 금지
- 스마트 크롭 3단계 폴백 체인 유지
- 이미지 방향 감지 로직 유지

### Shell 스크립트 (`widen-gracefully-memory.sh`)

**대안**: 간단한 사용/수정 필요 시

**실행 방법**:
```bash
# 단일 이미지
./widen-gracefully-memory.sh 입력파일.jpg

# 폴더 전체 (권장)
./widen-gracefully-memory.sh /path/to/folder
./widen-gracefully-memory.sh .  # 현재 폴더
```

**중요 규칙**:
1. 실행 권한 필요: `chmod +x`
2. 의존성: ImageMagick
3. 결과물 저장 경로: `shell/` 폴더
4. 임시 파일: 처리 후 자동 삭제

**코드 수정 시 확인사항**:
- 설정값 변수 확인 (`WATERMARK_WIDTH_RATIO` 등)
- ImageMagick 명령어 옵션 검증
- 파일 I/O 최소화 고려
- bash 문법 검증

## 처리 순서 (절대 변경 금지)

### Step 1: 워터마크 제거
- 위치: 하단 좌측
- 크기: 너비 23%, 높이 4%
- 방법: Gaussian Blur (radius=20)

### Step 2: 여백 제거 및 리사이즈
- 여백 감지: 5% fuzz 기반
- 목표 높이: 2160px
- 색감: ICC 프로필 유지 필수

### Step 3: 16:9 변환
- 목표 크기: 3840x2160
- 엣지 비율: 좌우 1%
- 블러 강도: radius=50

## 절대 금지사항

### Python 스크립트
- ❌ 파일 I/O 사용 (임시 파일 생성 금지)
- ❌ ICC 프로필 제거 (`-strip` 금지)
- ❌ 화질 저하 (`quality < 95` 금지)
- ❌ 클래스 구조 변경
- ❌ 스마트 크롭 3단계 폴백 체인 삭제
- ❌ 이미지 방향 감지 로직 삭제

### Shell 스크립트
- ❌ 영구적 임시 파일 남기기
- ❌ ICC 프로필 제거
- ❌ 처리 순서 변경

## 파일 경로 규칙

### 입력 파일
- 패턴: `ariel-introduction-{번호}-10000px.jpg`
- 예: `ariel-introduction-04-10000px.jpg`

### 출력 파일
- Python: `python/ariel-introduction-{번호}-4k.jpg`
- Shell: `shell/ariel-introduction-{번호}-4k.jpg`

### 임시 파일
- Python: 생성 금지 (메모리 처리)
- Shell: 자동 삭제 필수

## 성능 고려사항

### Python
- 메모리 처리 기본
- 병렬 처리 고려 시 최대 4개
- DecompressionBombWarning 무시

### Shell
- 파이프라인 최적화
- 임시 파일 최소화
- 병렬 처리 불권장 (성능 저하)

## 디버깅 규칙

### 에러 발생 시 확인사항
1. 입력 파일 경로 정확성
2. 결과 폴더 존재 여부 (`python/`, `shell/`)
3. 의존성 설치 여부
4. 버전 호환성

### 테스트 절차
1. 단일 이미지 테스트 필수
2. 결과 크기 확인 (3840x2160)
3. 색감 확인 (ICC 프로필 유지)
4. 대량 처리 전 샘플 테스트

## 수정 가이드

### 새로운 기능 추가
1. Python: `ImageConverter` 클래스에 새 메서드 추가
2. Shell: 새로운 bash 함수 추가
3. `process_image()` 또는 `main()` 함수 호출 순서 확인

### 파라미터 변경
1. `watermark_width_ratio`: 0.23 (기본값)
2. `watermark_height_ratio`: 0.04 (기본값)
3. `target_width`: 3840 (기본값)
4. `target_height`: 2160 (기본값)
5. `edge_percent`: 1 (기본값)

## 버전 관리

### Python 패키지
```bash
# 업데이트
uv pip install --upgrade Pillow numpy

# 버전 확인
uv pip list
```

### 스크립트 버전
- Python: 함수형 구조로 유지보수
- Shell: 단일 함수로 통합

## 권장 사용 패턴

### 단일 이미지
```bash
# Python 사용 (빠름)
python3 widen_gracefully.py 이미지.jpg
```

### 폴더 전체 처리 (권장)
```bash
# Python 사용
python3 widen_gracefully.py /path/to/folder

# 현재 폴더의 모든 이미지 처리
python3 widen_gracefully.py .
```

### 디버깅 필요한 경우
```bash
# Shell 사용 (간단한 로그)
./widen-gracefully-memory.sh /path/to/folder
```

### 병렬 처리 (대규모 폴더)
```bash
# 여러 폴더를 병렬로 처리 (최대 4개)
find /path/to/images -type d -maxdepth 1 | head -4 | xargs -P 4 -I {} python3 widen_gracefully.py "{}"
```

## 긴급 상황 대응

### 파일이 생성되지 않을 때
1. 결과 폴더 생성 확인
2. 디스크 공간 확인
3. 파일 권한 확인

### 색감이 이상할 때
1. ICC 프로필 코드 확인
2. `-strip` 옵션 제거 확인
3. 원본 이미지 색상 프로필 확인

### 성능 저하 시
1. 메모리 처리 확인 (Python)
2. 임시 파일 최소화 (Shell)
3. 병렬 처리 개수 줄이기

## 보안 규칙

### 이미지 파일
- 대용량 이미지 DecompressionBombWarning 무시
- 악성 이미지 처리 금지 (신뢰할 수 있는 소스만)

### 코드 보안
- 외부 라이브러리 버전 확인
- 쉘 인젝션 방지 (Shell)