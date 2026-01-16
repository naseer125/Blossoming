#!/bin/bash

# 우아하게 넓히기 스크립트
# 세로 이미지를 가로 16:9로 우아하게 변환

# 설정값
WATERMARK_WIDTH_RATIO=0.23
WATERMARK_HEIGHT_RATIO=0.04
TARGET_WIDTH=3840
TARGET_HEIGHT=2160
EDGE_PERCENT=1

# 워터마크 제거 함수
remove_watermark() {
    local input_file="$1"
    local output_file="$2"
    
    local img_width=$(identify -format "%w" "$input_file")
    local img_height=$(identify -format "%h" "$input_file")
    
    local watermark_width=$((img_width * 23 / 100))
    local watermark_height=$((img_height * 4 / 100))
    local watermark_y=$((img_height - watermark_height))
    
    echo "원본 크기: ${img_width}x${img_height}"
    echo "워터마크 영역: ${watermark_width}x${watermark_height} (좌하단)"
    
    magick "$input_file" \
        \( +clone -crop "${watermark_width}x${watermark_height}+0+${watermark_y}" -blur 0x20 \) \
        -compose over -geometry "+0+${watermark_y}" -composite \
        "$output_file"
}

# 여백 제거 및 리사이즈 함수
trim_and_resize() {
    local input_file="$1"
    local output_file="$2"
    
    magick "$input_file" -fuzz 5% -trim -resize x${TARGET_HEIGHT} \
        -quality 98 -sampling-factor 4:2:2 "$output_file"
    
    local resize_width=$(identify -format "%w" "$output_file")
    local resize_height=$(identify -format "%h" "$output_file")
    
    echo "리사이즈 후 크기: ${resize_width}x${resize_height}"
    echo "엣지 블러 폭: $((resize_width * EDGE_PERCENT / 100))px"
}

# 16:9 변환 함수
convert_to_16x9() {
    local input_file="$1"
    local output_file="$2"
    
    local resize_width=$(identify -format "%w" "$input_file")
    local resize_height=$(identify -format "%h" "$input_file")
    local edge_width=$((resize_width * EDGE_PERCENT / 100))
    local blur_width=$(( (TARGET_WIDTH - resize_width) / 2 ))
    
    # 좌우 엣지 추출
    echo "좌우 ${EDGE_PERCENT}% 엣지 추출 및 블러처리..."
    magick "$input_file" -crop "${edge_width}x${resize_height}+0+0" +repage left-edge.jpg
    magick "$input_file" -crop "${edge_width}x${resize_height}+$((resize_width - edge_width))+0" +repage right-edge.jpg
    
    # 블러 처리
    magick left-edge.jpg -resize "${blur_width}x${TARGET_HEIGHT}!" -blur 0x50 left-blur.jpg
    magick right-edge.jpg -resize "${blur_width}x${TARGET_HEIGHT}!" -blur 0x50 right-blur.jpg
    
    # 합치기
    echo "이미지 합치기..."
    magick left-blur.jpg "$input_file" right-blur.jpg +append \
        -quality 98 -sampling-factor 4:2:2 "$output_file"
}

# 임시 파일 삭제 함수
cleanup() {
    rm -f temp_watermark.jpg temp_resize.jpg left-edge.jpg right-edge.jpg left-blur.jpg right-blur.jpg
}

# 메인 함수
main() {
    if [ -z "$1" ]; then
        echo "Usage: $0 <image_file>"
        exit 1
    fi
    
    local input_file="$1"
    local basename=$(basename "$input_file" .jpg)
    local output_name=$(echo "$basename" | sed 's/-10000px//')-4k
    local output_path="shell/$output_name.jpg"
    
    echo "=== 우아하게 넓히기 ==="
    
    echo ""
    echo "=== Step 1: 워터마크 제거 ==="
    remove_watermark "$input_file" "temp_watermark.jpg"
    
    echo ""
    echo "=== Step 2: 여백 제거 및 리사이즈 ==="
    trim_and_resize "temp_watermark.jpg" "temp_resize.jpg"
    
    echo ""
    echo "=== Step 3: 16:9 변환 ==="
    convert_to_16x9 "temp_resize.jpg" "$output_path"
    
    cleanup
    
    local result_size=$(identify -format "%wx%h" "$output_path")
    echo ""
    echo "=== 완료! ==="
    echo "생성된 파일: shell/$output_name.jpg ($result_size)"
}

# 메인 함수 실행
main "$@"