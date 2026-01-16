#!/usr/bin/env python3

import os
import sys
from PIL import Image, ImageFilter, ImageChops
import numpy as np


class ImageConverter:
    def __init__(self):
        self.watermark_width_ratio = 0.23
        self.watermark_height_ratio = 0.04
        self.target_width = 3840
        self.target_height = 2160
        self.edge_percent = 1
        self.trim_fuzz = 5  # 5%
        self.quality = 98

    def remove_watermark(self, img):
        """워터마크 영역 블러처리 (메모리 처리)"""
        img_width, img_height = img.size

        watermark_width = int(img_width * self.watermark_width_ratio)
        watermark_height = int(img_height * self.watermark_height_ratio)
        watermark_y = img_height - watermark_height

        print(f"원본 크기: {img_width}x{img_height}")
        print(f"워터마크 영역: {watermark_width}x{watermark_height} (좌하단)")

        # 워터마크 영역 추출 및 블러처리
        watermark_region = img.crop((0, watermark_y, watermark_width, img_height))
        blurred_watermark = watermark_region.filter(ImageFilter.GaussianBlur(radius=20))

        # 원본 이미지에 블러처리된 워터마크 영역 합성
        img.paste(blurred_watermark, (0, watermark_y))

        return img

    def trim_whitespace(self, img):
        """여백 제거 (fuzz 기반, 메모리 처리)"""
        width, height = img.size

        # 이미지를 numpy 배열로 변환
        pixels = np.array(img)

        # 모서리 픽셀을 기준으로 배경색 결정 (상단 좌측 10x10 영역)
        background_color = np.mean(pixels[0:10, 0:10, :])

        # fuzz 퍼센트를 RGB 값으로 변환 (5% = 12.75/255)
        fuzz_threshold = 12.75

        # 상단 여백 감지
        top = 0
        for y in range(height):
            row_mean = np.mean(np.abs(pixels[y, :, :] - background_color))
            if row_mean > fuzz_threshold:  # 배경과 다른 색상 발견
                top = y
                break

        # 하단 여백 감지
        bottom = height
        for y in range(height - 1, -1, -1):
            row_mean = np.mean(np.abs(pixels[y, :, :] - background_color))
            if row_mean > fuzz_threshold:  # 배경과 다른 색상 발견
                bottom = y + 1
                break

        print(f"여백 제거: 상단 {top}px, 하단 {height - bottom}px")

        # 크롭 (새로운 이미지 객체 반환)
        trimmed_img = img.crop((0, top, width, bottom))

        return trimmed_img

    def resize_to_height(self, img):
        """높이를 target_height로 리사이즈 (메모리 처리)"""
        original_width, original_height = img.size

        new_width = int(original_width * (self.target_height / original_height))
        resized_img = img.resize((new_width, self.target_height), Image.LANCZOS)

        print(f"리사이즈 완료: {new_width}x{self.target_height}")
        return resized_img

    def convert_to_16x9(self, img):
        """16:9 변환 (좌우 블러처리로 캔버스 확장, 메모리 처리)"""
        current_width, current_height = img.size

        edge_width = int(current_width * self.edge_percent / 100)
        blur_width = (self.target_width - current_width) // 2

        print(f"현재 크기: {current_width}x{current_height}")
        print(f"엣지 블러 폭: {edge_width}px")

        # 좌측 엣지 추출
        left_edge = img.crop((0, 0, edge_width, current_height))
        left_edge_resized = left_edge.resize(
            (blur_width, self.target_height), Image.LANCZOS
        )
        left_blurred = left_edge_resized.filter(ImageFilter.GaussianBlur(radius=50))

        # 우측 엣지 추출
        right_edge = img.crop(
            (current_width - edge_width, 0, current_width, current_height)
        )
        right_edge_resized = right_edge.resize(
            (blur_width, self.target_height), Image.LANCZOS
        )
        right_blurred = right_edge_resized.filter(ImageFilter.GaussianBlur(radius=50))

        # 이미지 높이 조정
        if current_height != self.target_height:
            img = img.resize((current_width, self.target_height), Image.LANCZOS)

        # 이미지 합성
        result = Image.new("RGB", (self.target_width, self.target_height))
        result.paste(left_blurred, (0, 0))
        result.paste(img, (blur_width, 0))
        result.paste(right_blurred, (blur_width + current_width, 0))

        return result

    def process_image(self, input_path):
        """전체 처리 프로세스 (메모리 처리)"""
        if not os.path.exists(input_path):
            print(f"파일이 존재하지 않습니다: {input_path}")
            return

        basename = os.path.basename(input_path)
        name_without_ext = os.path.splitext(basename)[0]
        output_name = name_without_ext.replace("-10000px", "") + "-4k.jpg"
        output_path = os.path.join("python", output_name)

        os.makedirs("python", exist_ok=True)

        print("=== 우아하게 넓히기 ===")
        print("")

        try:
            with Image.open(input_path) as img:
                icc_profile = img.info.get("icc_profile")

                # Step 1: 워터마크 제거
                print("=== Step 1: 워터마크 제거 ===")
                img = self.remove_watermark(img)
                print("")

                # Step 2: 여백 제거 및 리사이즈
                print("=== Step 2: 여백 제거 및 리사이즈 ===")
                img = self.trim_whitespace(img)
                img = self.resize_to_height(img)
                print("")

                # Step 3: 16:9 변환
                print("=== Step 3: 16:9 변환 ===")
                result_img = self.convert_to_16x9(img)
                print("")

                # 결과 저장 (마지막에만 파일 I/O)
                if icc_profile:
                    result_img.save(
                        output_path,
                        quality=self.quality,
                        subsampling=1,
                        icc_profile=icc_profile,
                    )
                else:
                    result_img.save(output_path, quality=self.quality, subsampling=1)

                result_size = result_img.size
                print("=== 완료! ===")
                print(
                    f"생성된 파일: python/{output_name} ({result_size[0]}x{result_size[1]})"
                )

        except Exception as e:
            print(f"오류 발생: {e}")
            raise


def main():
    if len(sys.argv) < 2:
        print("Usage: python3 widen_gracefully.py <image_file_or_folder>")
        sys.exit(1)

    input_path = sys.argv[1]
    converter = ImageConverter()

    if os.path.isfile(input_path):
        converter.process_image(input_path)
    elif os.path.isdir(input_path):
        supported_extensions = (".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".tif")
        processed = 0
        for filename in sorted(os.listdir(input_path)):
            if filename.lower().endswith(supported_extensions):
                filepath = os.path.join(input_path, filename)
                print(f"\n{'=' * 60}")
                print(f"Processing: {filename}")
                print("=" * 60)
                converter.process_image(filepath)
                processed += 1
        print(f"\n총 {processed}개 이미지 처리 완료")
    else:
        print(f"오류: '{input_path}'은(는) 존재하지 않는 파일 또는 폴더입니다.")
        sys.exit(1)


if __name__ == "__main__":
    main()
