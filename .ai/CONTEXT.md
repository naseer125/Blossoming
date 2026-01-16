# AI Context

## 프로젝트 배경

이 프로젝트는 세로 방향의 고해상도 이미지를 가로 16:9 (3840x2160) 비율로 변환하는 이미지 처리 도구입니다. 원본 이미지의 워터마크 제거, 여백 제거, 그리고 자연스러운 캔버스 확장을 통해 화질 손상 없이 넓은 화면에 적합한 이미지를 생성합니다.

## 개발 과정

1. **초기 접근**: ImageMagick을 사용한 Bash 스크립트로 시작
2. **기능 확장**: 워터마크 제거, trim, 16:9 변환 등 단계별 기능 개발
3. **Python 도입**: Pillow 라이브러리 사용으로 성능 향상 (7.4배 빨라짐)
4. **메모리 최적화**: 임시 파일 사용 최소화로 처리 속도 개선
5. **최종 비교**: Python 4.781초 vs Shell 52.468초 (5개 이미지 기준)

## 기술 스택

### Python 버전
- **라이브러리**: Pillow (이미지 처리), numpy (수치 연산)
- **패키지 매니저**: uv
- **가상환경**: `/Users/crong/Pictures/Wallpaper/.venv`
- **파이썬 버전**: 3.13.11

### Shell 버전
- **도구**: ImageMagick
- **언어**: Bash
- **특징**: 간단한 설치, 범용성

## 주요 기술적 결정

### Python 선택 이유
1. **성능**: 11배 더 빠른 처리 속도
2. **메모리 처리**: 전체 과정을 메모리 내에서 처리
3. **색감 보존**: ICC 프로필 자동 유지
4. **유지보수성**: 함수형 구조로 코드 관리 용이

### 메모리 최적화 접근
- Python: 이미 기본적으로 메모리 처리 (41% 성능 향상)
- Shell: ImageMagick 파이프라인 활용 (12% 성능 향상)

## 프로젝트 구조

```
temp/
├── .ai/                      # AI 설정 폴더
│   ├── AI.ignore             # AI 무시 파일 목록
│   ├── RULES.md              # AI 규칙
│   └── CONTEXT.md            # 본 파일
├── .venv/                    # uv 가상환경 (Python 3.13.11)
├── python/                   # Python 결과물 폴더
├── shell/                    # Shell 결과물 폴더
├── *.jpg                     # 입력 이미지 파일들
├── widen_gracefully.py      # Python 메인 스크립트
├── widen-gracefully-memory.sh # Shell 메인 스크립트
├── embrace-horizon.sh       # 이전 버전 (참고용)
├── remove-watermark-trim-hq.sh # 이전 버전 (참고용)
├── grace.md                  # 프로젝트 기록
└── README.md                 # 프로젝트 문서
```

## 이미지 처리 알고리즘

### 1단계: 워터마크 제거
- **위치**: 하단 좌측
- **크기**: 너비 23%, 높이 4%
- **방법**: Gaussian Blur (radius=20)
- **목표**: 워터마크를 자연스럽게 흐리게 처리

### 2단계: 여백 제거
- **방법**: 배경색 기반 감지 (5% fuzz)
- **알고리즘**: 
  - 상단 좌측 10x10 영역을 배경색으로 설정
  - 상하 행을 스캔하며 배경색과 다른 영역 감지
  - 감지된 영역 외부를 크롭
- **목표**: 불필요한 여백 제거

### 3단계: 리사이즈
- **목표 높이**: 2160px
- **방법**: LANCZOS 리샘플링 (고품질)
- **특징**: 너비 비율 유지

### 4단계: 16:9 변환
- **목표 크기**: 3840x2160
- **방법**:
  - 좌우 1% 엣지 추출
  - 엣지를 확장 후 Gaussian Blur (radius=50)
  - 중앙 이미지 + 좌우 블러 이미지 합성
- **목표**: 자연스러운 캔버스 확장

## 색상 처리

### ICC 프로필 보존
```python
icc_profile = img.info.get('icc_profile')
# 처리 후
if icc_profile:
    img.save(output, icc_profile=icc_profile)
```

### 중요성
- 원본 색감 유지
- 다른 디스플레이에서 색상 재현 보장
- 색감 변화 방지

## 성능 데이터

### 처리 속도 (5개 이미지)
| 버전 | 총 시간 | 평균/파일 | 비율 |
|------|---------|-----------|------|
| Python 메모리 | 4.781초 | 0.956초 | 1x |
| Python 파일 I/O | 8.092초 | 1.618초 | 1.7x 느림 |
| Shell 메모리 | 52.468초 | 10.494초 | 11x 느림 |
| Shell 파일 I/O | 59.64초 | 11.928초 | 12.4x 느림 |

### 최적화 효과
- Python: 메모리 처리로 41% 향상
- Shell: 파이프라인 최적화로 12% 향상

## 사용 시나리오

### 단일 이미지 변환
```bash
python3 widen_gracefully.py input.jpg
# 결과: python/input-4k.jpg (3840x2160)
```

### 대량 처리
```bash
for file in *.jpg; do
    python3 widen_gracefully.py "$file"
done
```

### 배치 처리 (병렬)
```bash
# GNU parallel 사용
ls *.jpg | parallel -j 4 python3 widen_gracefully.py {}
```

## 알려진 문제

### DecompressionBombWarning
- **원인**: PIL의 대용량 이미지 보안 경고
- **해결**: 무시해도 됨 (안전한 이미지 처리)
- **발생 조건**: 89MB 이상 이미지

### 파일 명명
- **패턴**: `ariel-introduction-{번호}-10000px.jpg`
- **결과**: `ariel-introduction-{번호}-4k.jpg`
- **위치**: `python/` 또는 `shell/` 폴더

## 확장 가능성

### 새로운 비율 지원
```python
# 설정값 수정
target_width = 1920  # 예: 1080p
target_height = 1080
```

### 다른 워터마크 위치
```python
# 워터마크 좌표 수정
watermark_x = img_width - watermark_width  # 우측
watermark_y = img_height - watermark_height  # 하단
```

### 추가 전처리
- 대비 조정
- 색상 보정
- 노이즈 제거

## 유지보수 가이드

### Python 스크립트 수정
1. `ImageConverter` 클래스 확인
2. 메서드 시그니처 유지
3. ICC 프로필 보존 확인
4. 메모리 처리 유지
5. uv 환경에서 테스트

### Shell 스크립트 수정
1. 변수 설정 확인
2. ImageMagick 명령어 검증
3. 파일 I/O 최소화
4. bash 문법 검증
5. 테스트 이미지로 확인

## 테스트 절차

### 기능 테스트
```bash
# 1. 단일 이미지
python3 widen_gracefully.py test.jpg
identify python/test-4k.jpg

# 2. 여백 제거 확인
# - 원본: 8708x11608
# - 결과: 3840x2160

# 3. 워터마크 제거 확인
# - 하단 좌측 워터마크 사라짐
```

### 성능 테스트
```bash
time python3 widen_gracefully.py test.jpg
# 예상: < 1초
```

### 색감 테스트
```bash
# 원본과 결과 색상 비교
compare test.jpg python/test-4k.jpg diff.jpg
```

## 프로젝트 목표

### 현재 상태
- ✅ 워터마크 제거
- ✅ 여백 제거
- ✅ 16:9 변환
- ✅ 색감 보존
- ✅ 메모리 최적화
- ✅ 성능 최적화

### 향후 계획
- 병렬 처리 지원
- 다른 비율 지원
- GUI 도구 개발
- 웹 인터페이스

## 참고 자료

### 이미지 처리
- Pillow 공식 문서: https://pillow.readthedocs.io/
- ImageMagick 공식: https://imagemagick.org/

### 성능 최적화
- Python 프로파일링: `cProfile`, `timeit`
- ImageMagick 프로파일링: `magick -debug all`

### 색상 관리
- ICC 프로필: https://www.color.org/
- sRGB 공간: https://en.wikipedia.org/wiki/SRGB