#!/bin/bash

# 우아하게 넓히기 스크립트 (메모리 최적화 버전)
# 세로 이미지를 가로 16:9로 우아하게 변환 - 최소 파일 I/O

# 설정값
WATERMARK_WIDTH_RATIO=0.23
WATERMARK_HEIGHT_RATIO=0.04
TARGET_WIDTH=3840
TARGET_HEIGHT=2160
EDGE_PERCENT=1

process_image() {
    local input_file="$1"
    local basename=$(basename "$input_file")
    basename="${basename%.*}"
    local output_name=$(echo "$basename" | sed 's/-10000px//')-4k
    local output_path="shell/$output_name.jpg"

    echo "=== 우아하게 넓히기 (메모리 최적화) ==="
    echo ""

    # 원본 이미지 크기 확인
    local img_width=$(identify -format "%w" "$input_file")
    local img_height=$(identify -format "%h" "$input_file")

    # 워터마크 영역 계산
    local watermark_width=$((img_width * 23 / 100))
    local watermark_height=$((img_height * 4 / 100))
    local watermark_y=$((img_height - watermark_height))

    echo "=== Step 1: 워터마크 제거, 여백 제거, 리사이즈 ==="
    echo "원본 크기: ${img_width}x${img_height}"
    echo "워터마크 영역: ${watermark_width}x${watermark_height} (좌하단)"

    # Step 1: 워터마크 제거 + 여백 제거 + 리사이즈를 하나의 파이프라인으로 처리
    magick "$input_file" \
        \( +clone -crop "${watermark_width}x${watermark_height}+0+${watermark_y}" -blur 0x20 \) \
        -compose over -geometry "+0+${watermark_y}" -composite \
        -fuzz 5% -trim -resize x${TARGET_HEIGHT} \
        -quality 98 -sampling-factor 4:2:2 \
        "temp_step1.jpg"

    # 리사이즈된 크기 확인
    local resize_width=$(identify -format "%w" "temp_step1.jpg")
    local resize_height=$(identify -format "%h" "temp_step1.jpg")
    local edge_width=$((resize_width * EDGE_PERCENT / 100))
    local blur_width=$(( (TARGET_WIDTH - resize_width) / 2 ))
    local blur_remainder=$(( (TARGET_WIDTH - resize_width) % 2 ))

    echo "리사이즈 후 크기: ${resize_width}x${resize_height}"
    echo "엣지 블러 폭: ${edge_width}px"
    echo ""

    echo "=== Step 2: 16:9 변환 (좌우 블러처리) ==="
    echo "좌우 ${EDGE_PERCENT}% 엣지 추출 및 블러처리..."

    # 좌측 엣지 처리
    magick "temp_step1.jpg" \
        -crop "${edge_width}x${resize_height}+0+0" +repage \
        -resize "${blur_width}x${TARGET_HEIGHT}!" -blur 0x50 \
        "left-blur.jpg"

    # 우측 엣지 처리 (나머지 픽셀을 우측에 추가)
    local right_blur_width=$(( blur_width + blur_remainder ))
    magick "temp_step1.jpg" \
        -crop "${edge_width}x${resize_height}+$((resize_width - edge_width))+0" +repage \
        -resize "${right_blur_width}x${TARGET_HEIGHT}!" -blur 0x50 \
        "right-blur.jpg"

    # 합치기
    echo "이미지 합치기..."
    magick "left-blur.jpg" "temp_step1.jpg" "right-blur.jpg" +append \
        -quality 98 -sampling-factor 4:2:2 \
        "$output_path"

    # 임시 파일 삭제
    rm -f temp_step1.jpg left-blur.jpg right-blur.jpg

    # 결과 확인
    local result_size=$(identify -format "%wx%h" "$output_path")
    echo ""
    echo "=== 완료! ==="
    echo "생성된 파일: shell/$output_name.jpg ($result_size)"
}

# 메인 함수
main() {
    if [ -z "$1" ]; then
        echo "Usage: $0 <image_file_or_folder>"
        exit 1
    fi

    local input_path="$1"

    if [ -f "$input_path" ]; then
        process_image "$input_path"
    elif [ -d "$input_path" ]; then
        local processed=0
        for file in "$input_path"/*.{jpg,jpeg,JPG,JPEG,png,bmp,tif,tiff}; do
            if [ -f "$file" ]; then
                echo ""
                echo "============================================================"
                echo "Processing: $(basename "$file")"
                echo "============================================================"
                process_image "$file"
                processed=$((processed + 1))
            fi
        done
        echo ""
        echo "총 ${processed}개 이미지 처리 완료"
    else
        echo "오류: '$input_path'은(는) 존재하지 않는 파일 또는 폴더입니다."
        exit 1
    fi
}

# 메인 함수 실행
main "$@"