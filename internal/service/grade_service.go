package service

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/model"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/storage"
)

var ErrInProgress = errors.New("❗️ already in progress")

type GradeService struct {
	isRunning      atomic.Bool
	LastTimeParsed time.Time

	csvWriter *storage.CSVwriter
	fetcher   *MoodleFetcher

	badTitles []string
}

func NewGradeService(fetcher *MoodleFetcher, csvWriter *storage.CSVwriter, badTitles []string) *GradeService {
	return &GradeService{
		fetcher:   fetcher,
		csvWriter: csvWriter,
		badTitles: badTitles,
	}
}

func (p *GradeService) WaitTillFinishedRunning() {
	for p.isRunning.Load() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (p *GradeService) ParseAndCompare() ([]model.Change, error) {
	if !p.isRunning.CompareAndSwap(false, true) {
		slog.Debug("ParseAndCompare:already_running")
		return nil, ErrInProgress
	}
	defer p.isRunning.Store(false)

	if err := p.fetcher.IsLogined(); err != nil {
		err = p.fetcher.Login()
		if err != nil {
			slog.Error("Login failed", "error", err)
		}
	}

	buf, err := p.fetcher.GetGradesPage()
	if err != nil {
		return nil, err
	}

	links, err := extractGradesLinks(buf, p.badTitles)
	if err != nil {
		return nil, err
	}
	slog.Debug("Successfully extracted links", "len", len(links))

	var wg sync.WaitGroup
	var mux sync.Mutex
	var TotalChanges []model.Change
	for _, link := range links {
		slog.Debug("Processing link", "link", link)
		wg.Go(func() {
			buf, err := p.fetcher.Fetch(link)
			if err != nil {
				slog.Error("Failed to fetch grade page", "link", link, "error", err)
			}

			courseName, newItems, err := extractItems(buf)
			if err != nil {
				slog.Error("Failed to extract items", "link", link, "error", err)
				return
			}

			oldItems, err := p.readItemsCourse(courseName)
			exists := !errors.Is(err, os.ErrNotExist)
			if err != nil && exists {
				slog.Error("Failed to read old items", "course", courseName, "error", err)
				return
			}

			if exists {
				CourseChanges := Compare(courseName, oldItems, newItems)
				slog.Debug("Course changes found", "course", courseName, "count", len(CourseChanges))

				mux.Lock()
				TotalChanges = append(TotalChanges, CourseChanges...)
				mux.Unlock()
			}

			err = p.writeItems(courseName, newItems)
			if err != nil {
				slog.Error("Failed to write new items", "course", courseName, "error", err)
			}
		})
	}

	wg.Wait()

	p.LastTimeParsed = time.Now()
	slog.Debug("ParseAndCompare:done", "total_changes", len(TotalChanges))
	return TotalChanges, nil
}

func (p *GradeService) GetLastTimeParsed() time.Time {
	slog.Debug("GetLastTimeParsed", "last", p.LastTimeParsed)
	return p.LastTimeParsed
}
func buildFilePath(courseName string) string {
	path := sanitizeFilename(courseName) + "_grades.csv"
	return path
}

func sanitizeFilename(name string) string {
	// Replace common invalid filename characters with underscores
	invalidChars := []string{" ", ":", "/", "\\", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}
	// Trim whitespace and limit length to avoid issues
	name = strings.TrimSpace(name)
	if len(name) > 255 {
		name = name[:255]
	}
	slog.Debug("sanitizeFilename", "result", name)
	return name
}

func (p *GradeService) writeItems(courseName string, items []*model.GradeRow) error {
	slog.Debug("writeItems", "course", courseName, "items", len(items))

	record := make([][]string, 0, len(items))
	for _, item := range items {
		record = append(record, item.ToStringSlice())
	}

	return p.csvWriter.Write(buildFilePath(courseName), record)
}

func (p *GradeService) readItemsCourse(courseName string) ([]*model.GradeRow, error) {
	slog.Debug("readItemsCourse", "course", courseName)

	return p.readItemsFile(buildFilePath(courseName))
}

func (p *GradeService) readItemsFile(courseFilePath string) ([]*model.GradeRow, error) {
	slog.Debug("readItems", "course", courseFilePath)

	records, err := p.csvWriter.Read(courseFilePath)
	if err != nil {
		return nil, err
	}

	var rows []*model.GradeRow
	for _, record := range records {
		rows = append(rows, model.NewGradeRow(record))
	}

	return rows, nil
}

func (p *GradeService) GetCourseNamesList() ([]string, error) {
	files, err := p.csvWriter.ListFiles()
	if err != nil {
		return nil, err
	}
	slog.Debug("GetCourseNamesList", "files", len(files))

	return files, nil
}

func (p *GradeService) GetCourseGrades(courseName string) ([]*model.GradeRow, error) {
	slog.Debug("GetCourseGrades", "course", courseName)
	return p.readItemsCourse(courseName)
}

func (p *GradeService) GetCourseFile(coursefile string) ([]*model.GradeRow, error) {
	slog.Debug("GetCourseFile", "course", coursefile)
	return p.readItemsFile(coursefile)
}

func Compare(courseName string, old, new []*model.GradeRow) []model.Change {
	slog.Debug("Compare:start", "course", courseName, "old", len(old), "new", len(new))
	mp := map[string]*model.GradeRow{}
	for _, s := range old {
		mp[s.AssName] = s
	}

	var changes []model.Change
	for _, s := range new {
		old, ok := mp[s.AssName]
		if !ok {
			changes = append(changes, model.Change{
				CourseName: courseName,
				TP:         model.NewElement,
				New:        s,
			})
		} else {
			if !old.IsEqual(s) {
				changes = append(changes, model.Change{
					CourseName: courseName,
					TP:         model.Changed,
					Old:        old,
					New:        s,
				})
			}
		}
	}

	return changes
}
