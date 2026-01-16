package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
)

// ImageConverter 이미지 변환기
type ImageConverter struct {
	WatermarkWidthRatio  float64
	WatermarkHeightRatio float64
	TargetWidth           int
	TargetHeight          int
	EdgePercent           int
	TrimFuzz              int     // 5% = 12.75/255 (0-255 스케일)
	Quality               int     // JPEG quality
}

// NewImageConverter 새로운 이미지 변환기 생성
func NewImageConverter() *ImageConverter {
	return &ImageConverter{
		WatermarkWidthRatio:  0.23,
		WatermarkHeightRatio: 0.04,
		TargetWidth:           3840,
		TargetHeight:          2160,
		EdgePercent:           1,
		TrimFuzz:              13, // 약 5%
		Quality:               98,
	}
}

// ProcessImage 이미지 처리
func (c *ImageConverter) ProcessImage(inputPath string) error {
	basename := filepath.Base(inputPath)
	ext := filepath.Ext(basename)
	nameWithoutExt := strings.TrimSuffix(basename, ext)

	// -10000px 제거하고 -4k 추가
	outputName := strings.Replace(nameWithoutExt, "-10000px", "", -1) + "-4k.jpg"
	outputPath := filepath.Join("go", outputName)

	fmt.Println("=== 우아하게 넓히기 ===")
	fmt.Println()

	// 출력 폴더 생성
	if err := os.MkdirAll("go", 0755); err != nil {
		return fmt.Errorf("출력 폴더 생성 실패: %w", err)
	}

	// 원본 이미지 열기
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer file.Close()

	img, err := imaging.Decode(file)
	if err != nil {
		return fmt.Errorf("이미지 디코딩 실패: %w", err)
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	fmt.Printf("원본 크기: %dx%d\n", width, height)

	// Step 1: 워터마크 제거
	fmt.Println("=== Step 1: 워터마크 제거 ===")
	img = c.removeWatermark(img)
	fmt.Println()

	// Step 2: 여백 제거 및 리사이즈
	fmt.Println("=== Step 2: 여백 제거 및 리사이즈 ===")
	img = c.trimWhitespace(img)
	img = c.resizeToHeight(img)
	fmt.Println()

	// Step 3: 16:9 변환
	fmt.Println("=== Step 3: 16:9 변환 ===")
	result := c.convertTo16x9(img)
	fmt.Println()

	// 결과 저장
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("출력 파일 생성 실패: %w", err)
	}
	defer outputFile.Close()

	// JPEG 저장 (ICC 프로필 보존은 별도 처리 필요)
	options := &jpeg.Options{Quality: c.Quality}
	if err := jpeg.Encode(outputFile, result, options); err != nil {
		return fmt.Errorf("이미지 저장 실패: %w", err)
	}

	resultBounds := result.Bounds()
	resultWidth, resultHeight := resultBounds.Dx(), resultBounds.Dy()

	fmt.Println("=== 완료! ===")
	fmt.Printf("생성된 파일: go/%s (%dx%d)\n", outputName, resultWidth, resultHeight)

	return nil
}

// removeWatermark 워터마크 영역 블러처리
func (c *ImageConverter) removeWatermark(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	watermarkWidth := int(float64(width) * c.WatermarkWidthRatio)
	watermarkHeight := int(float64(height) * c.WatermarkHeightRatio)
	watermarkY := height - watermarkHeight

	fmt.Printf("워터마크 영역: %dx%d (좌하단)\n", watermarkWidth, watermarkHeight)

	// 워터마크 영역 추출
	watermarkRegion := imaging.Crop(img, image.Rect(0, watermarkY, watermarkWidth, height))

	// 블러처리 (Gaussian Blur)
	blurredWatermark := imaging.Blur(watermarkRegion, 20)

	// 원본 이미지에 블러처리된 워터마크 영역 합성
	result := imaging.Clone(img)
	for y := watermarkY; y < height; y++ {
		for x := 0; x < watermarkWidth; x++ {
			result.Set(x, y, blurredWatermark.At(x, watermarkY))
		}
	}

	return result
}

// trimWhitespace 여백 제거
func (c *ImageConverter) trimWhitespace(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// 배경색 결정 (상단 좌측 10x10 영역 평균)
	var r, g, b, count uint32
	for y := 0; y < 10 && y < height; y++ {
		for x := 0; x < 10 && x < width; x++ {
			pr, pg, pb, _ := img.At(x, y).RGBA()
			r += pr >> 8
			g += pg >> 8
			b += pb >> 8
			count++
		}
	}

	bgR := float64(r) / float64(count)
	bgG := float64(g) / float64(count)
	bgB := float64(b) / float64(count)

	// 상단 여백 감지 (모든 열 스캔)
	top := 0
	fuzzThreshold := 12.75 // 5% = 12.75/255

	for y := 0; y < height; y++ {
		rowMean := 0.0
		sampleCount := 0

		// 성능을 위해 10픽셀 간격으로 표본 추출
		for x := 0; x < width; x += 10 {
			pr, pg, pb, _ := img.At(x, y).RGBA()
			rDiff := absFloat(float64(pr>>8) - bgR)
			gDiff := absFloat(float64(pg>>8) - bgG)
			bDiff := absFloat(float64(pb>>8) - bgB)
			rowMean += (rDiff + gDiff + bDiff) / 3
			sampleCount++
		}

		if sampleCount > 0 {
			rowMean /= float64(sampleCount)
		}

		if rowMean > fuzzThreshold {
			top = y
			break
		}
	}

	// 하단 여백 감지 (모든 열 스캔)
	bottom := height
	for y := height - 1; y >= 0; y-- {
		rowMean := 0.0
		sampleCount := 0

		// 성능을 위해 10픽셀 간격으로 표본 추출
		for x := 0; x < width; x += 10 {
			pr, pg, pb, _ := img.At(x, y).RGBA()
			rDiff := absFloat(float64(pr>>8) - bgR)
			gDiff := absFloat(float64(pg>>8) - bgG)
			bDiff := absFloat(float64(pb>>8) - bgB)
			rowMean += (rDiff + gDiff + bDiff) / 3
			sampleCount++
		}

		if sampleCount > 0 {
			rowMean /= float64(sampleCount)
		}

		if rowMean > fuzzThreshold {
			bottom = y + 1
			break
		}
	}

	fmt.Printf("여백 제거: 상단 %dpx, 하단 %dpx\n", top, height-bottom)

	// 크롭
	if top == 0 && bottom == height {
		return img
	}

	return imaging.Crop(img, image.Rect(0, top, width, bottom))
}

// resizeToHeight 높이를 target_height로 리사이즈
func (c *ImageConverter) resizeToHeight(img image.Image) image.Image {
	bounds := img.Bounds()
	originalWidth, originalHeight := bounds.Dx(), bounds.Dy()

	newWidth := int(float64(originalWidth) * (float64(c.TargetHeight) / float64(originalHeight)))
	resized := imaging.Resize(img, newWidth, c.TargetHeight, imaging.Lanczos)

	fmt.Printf("리사이즈 완료: %dx%d\n", newWidth, c.TargetHeight)

	return resized
}

// convertTo16x9 16:9 변환 (좌우 블러처리로 캔버스 확장)
func (c *ImageConverter) convertTo16x9(img image.Image) image.Image {
	bounds := img.Bounds()
	currentWidth, currentHeight := bounds.Dx(), bounds.Dy()

	edgeWidth := int(float64(currentWidth) * float64(c.EdgePercent) / 100)
	blurWidth := (c.TargetWidth - currentWidth) / 2

	fmt.Printf("현재 크기: %dx%d\n", currentWidth, currentHeight)
	fmt.Printf("엣지 블러 폭: %dpx\n", edgeWidth)

	// 높이 조정
	if currentHeight != c.TargetHeight {
		img = imaging.Resize(img, currentWidth, c.TargetHeight, imaging.Lanczos)
		currentHeight = c.TargetHeight
	}

	// 좌측 엣지 추출
	leftEdge := imaging.Crop(img, image.Rect(0, 0, edgeWidth, currentHeight))
	leftEdgeResized := imaging.Resize(leftEdge, blurWidth, c.TargetHeight, imaging.Lanczos)
	leftBlurred := imaging.Blur(leftEdgeResized, 50)

	// 우측 엣지 추출
	rightEdge := imaging.Crop(img, image.Rect(currentWidth-edgeWidth, 0, currentWidth, currentHeight))
	rightEdgeResized := imaging.Resize(rightEdge, blurWidth, c.TargetHeight, imaging.Lanczos)
	rightBlurred := imaging.Blur(rightEdgeResized, 50)

	// 이미지 높이 조정
	middle := imaging.Resize(img, currentWidth, c.TargetHeight, imaging.Lanczos)

	// 이미지 합성
	result := imaging.New(c.TargetWidth, c.TargetHeight, color.Black)

	// 좌측 블러
	drawImage(result, leftBlurred, 0, 0)

	// 중간 이미지
	drawImage(result, middle, blurWidth, 0)

	// 우측 블러
	drawImage(result, rightBlurred, blurWidth+currentWidth, 0)

	return result
}

// drawImage 이미지 그리기
func drawImage(dst *image.NRGBA, src image.Image, offsetX, offsetY int) {
	srcBounds := src.Bounds()
	for y := srcBounds.Min.Y; y < srcBounds.Max.Y; y++ {
		for x := srcBounds.Min.X; x < srcBounds.Max.X; x++ {
			dstX := offsetX + x - srcBounds.Min.X
			dstY := offsetY + y - srcBounds.Min.Y
			dst.Set(dstX, dstY, src.At(x, y))
		}
	}
}

// abs 절대값 (int)
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// absFloat 절대값 (float64)
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// main 메인 함수
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "사용법:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  widen-gracefully-go [플래그] <파일 또는 폴더>\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "플래그:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)
	converter := NewImageConverter()

	supportedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".tiff": true,
		".tif":  true,
	}

	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("오류: '%s'은(는) 존재하지 않는 파일 또는 폴더입니다.\n", inputPath)
		os.Exit(1)
	}

	if fileInfo.IsDir() {
		// 폴더 처리
		processed := 0
		files, err := os.ReadDir(inputPath)
		if err != nil {
			fmt.Printf("폴더 읽기 실패: %v\n", err)
			os.Exit(1)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filename := file.Name()
			ext := strings.ToLower(filepath.Ext(filename))
			if !supportedExtensions[ext] {
				continue
			}

			filepath := filepath.Join(inputPath, filename)
			fmt.Printf("\n%s\n", strings.Repeat("=", 60))
			fmt.Printf("Processing: %s\n", filename)
			fmt.Println(strings.Repeat("=", 60))

			if err := converter.ProcessImage(filepath); err != nil {
				fmt.Printf("오류: %v\n", err)
				continue
			}

			processed++
		}

		fmt.Printf("\n총 %d개 이미지 처리 완료\n", processed)
	} else {
		// 단일 파일 처리
		ext := strings.ToLower(filepath.Ext(inputPath))
		if !supportedExtensions[ext] {
			fmt.Printf("오류: 지원하지 않는 파일 형식입니다 (%s)\n", ext)
			os.Exit(1)
		}

		if err := converter.ProcessImage(inputPath); err != nil {
			fmt.Printf("오류: %v\n", err)
			os.Exit(1)
		}
	}
}
