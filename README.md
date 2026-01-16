# Blossoming

세로 이미지를 가로 16:9 비율로 우아하게 넓히는 도구

## 빠른 시작

```bash
# Python 버전 (권장)
python3 widen_gracefully.py 입력파일.jpg

# Shell 버전 (대안)
chmod +x widen-gracefully-memory.sh
./widen-gracefully-memory.sh 입력파일.jpg
```

## 이 도구가 하는 일

- **워터마크 제거**: 하단 워터마크를 블러처리로 자연스럽게 제거
- **여백 제거**: 상하 불필요한 여백을 자동으로 제거
- **16:9 변환**: 3840x2160 해상도로 자연스럽게 확장
- **색감 보존**: 원본 색상 그대로 유지

## 설치

### Python 버전

```bash
# 의존성 설치
uv pip install Pillow numpy
```

### Shell 버전

```bash
# 이미 ImageMagick이 설치되어 있다면 건너뜀
# 설치 확인
magick -version
```

## 사용법

### 단일 이미지 변환

```bash
# Python
python3 widen_gracefully.py ariel-introduction-04-10000px.jpg
# 결과: python/ariel-introduction-04-4k.jpg

# Shell
./widen-gracefully-memory.sh ariel-introduction-04-10000px.jpg
# 결과: shell/ariel-introduction-04-4k.jpg
```

### 여러 이미지 한번에 변환

```bash
# Python
for file in *.jpg; do
    python3 widen_gracefully.py "$file"
done

# Shell
for file in *.jpg; do
    ./widen-gracefully-memory.sh "$file"
done
```

## 결과물

### 폴더 구조

```
temp/
├── python/              # Python 결과물
└── shell/               # Shell 결과물
```

### 변환 예시

| 원본 | 결과 | 해상도 |
|------|------|--------|
| 8708x11608 | python/ariel-introduction-04-4k.jpg | 3840x2160 |
| 8708x11608 | shell/ariel-introduction-04-4k.jpg | 3840x2160 |

## 어떤 버전을 써야 하나요?

### Python 버전을 써야 할 때
- **빠른 처리가 필요할 때**: 11배 더 빠름
- **대량 이미지 처리**: 성능 차이 큼
- **일반적인 사용**: 대부분의 경우

### Shell 버전을 써야 할 때
- **간단한 설치를 원할 때**: ImageMagick만 있으면 됨
- **수정이 자주 필요할 때**: Bash 스크립트 쉬운 편집
- **설치 간편성**: 추가 의존성 없음

## 성능 비교

5개 이미지 처리 기준

| 버전 | 총 시간 | 평균/파일 |
|------|---------|-----------|
| Python | 4.781초 | 0.956초 |
| Shell | 52.468초 | 10.494초 |

**Python이 11배 더 빠릅니다**

## 자주 묻는 질문

### Q. 색감이 변한 것 같아요
A. ICC 프로필을 보존해서 원본 색감 그대로 유지됩니다.

### Q. "DecompressionBombWarning"이 떠요
A. PIL의 대용량 이미지 경고입니다. 무시해도 안전합니다.

### Q. 결과물이 어디에 생겨요?
A. Python은 `python/` 폴더, Shell은 `shell/` 폴더에 저장됩니다.

### Q. 해상도를 바꾸고 싶어요
A. 스크립트 내의 `target_width`와 `target_height` 값을 수정하세요.

## 라이선스

개인 프로젝트로 사용됩니다.

## 문제 신고

버그나 개선 제안은 환영합니다.