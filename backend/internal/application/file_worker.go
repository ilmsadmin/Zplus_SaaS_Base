package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// FileWorker handles background file processing tasks
type FileWorker struct {
	fileRepo    domain.FileRepository
	jobRepo     domain.FileProcessingJobRepository
	maxRetries  int
	retryDelay  time.Duration
	workerCount int
	jobChannel  chan *domain.FileProcessingJob
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewFileWorker creates a new file worker
func NewFileWorker(
	fileRepo domain.FileRepository,
	jobRepo domain.FileProcessingJobRepository,
) *FileWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &FileWorker{
		fileRepo:    fileRepo,
		jobRepo:     jobRepo,
		maxRetries:  3,
		retryDelay:  time.Second * 30,
		workerCount: 5,
		jobChannel:  make(chan *domain.FileProcessingJob, 100),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the file processing workers
func (w *FileWorker) Start() {
	log.Println("Starting file processing workers...")

	// Start worker goroutines
	for i := 0; i < w.workerCount; i++ {
		go w.worker(i)
	}

	// Start job scheduler
	go w.jobScheduler()

	log.Printf("Started %d file processing workers", w.workerCount)
}

// Stop stops the file processing workers
func (w *FileWorker) Stop() {
	log.Println("Stopping file processing workers...")
	w.cancel()
	close(w.jobChannel)
}

// SubmitJob submits a new job for processing
func (w *FileWorker) SubmitJob(job *domain.FileProcessingJob) error {
	select {
	case w.jobChannel <- job:
		return nil
	case <-w.ctx.Done():
		return fmt.Errorf("worker is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// worker processes jobs from the job channel
func (w *FileWorker) worker(workerID int) {
	log.Printf("Worker %d started", workerID)

	for {
		select {
		case job := <-w.jobChannel:
			if job == nil {
				log.Printf("Worker %d: received nil job, shutting down", workerID)
				return
			}

			log.Printf("Worker %d: processing job %s (type: %s)", workerID, job.ID, job.JobType)
			w.processJob(job)

		case <-w.ctx.Done():
			log.Printf("Worker %d: context cancelled, shutting down", workerID)
			return
		}
	}
}

// jobScheduler periodically checks for pending jobs
func (w *FileWorker) jobScheduler() {
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.schedulePendingJobs()
		case <-w.ctx.Done():
			log.Println("Job scheduler: context cancelled, shutting down")
			return
		}
	}
}

// schedulePendingJobs retrieves and schedules pending jobs
func (w *FileWorker) schedulePendingJobs() {
	jobs, err := w.jobRepo.FindByStatus(w.ctx, domain.JobStatusPending)
	if err != nil {
		log.Printf("Error retrieving pending jobs: %v", err)
		return
	}

	for _, job := range jobs {
		select {
		case w.jobChannel <- job:
			log.Printf("Scheduled pending job: %s", job.ID)
		default:
			log.Printf("Job queue full, skipping job: %s", job.ID)
			return
		}
	}
}

// processJob processes a single job
func (w *FileWorker) processJob(job *domain.FileProcessingJob) {
	// Update job status to processing
	job.Status = domain.JobStatusProcessing
	now := time.Now()
	job.ProcessingStartedAt = &now

	if err := w.jobRepo.Update(w.ctx, job); err != nil {
		log.Printf("Error updating job status: %v", err)
		return
	}

	var err error

	// Process based on job type
	switch job.JobType {
	case domain.JobTypeVirusScan:
		err = w.processVirusScan(job)
	case domain.JobTypeImageResize:
		err = w.processImageResize(job)
	case domain.JobTypeImageCrop:
		err = w.processImageCrop(job)
	case domain.JobTypeThumbnailGeneration:
		err = w.processThumbnailGeneration(job)
	case domain.JobTypeMetadataExtraction:
		err = w.processMetadataExtraction(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	// Update job status based on result
	if err != nil {
		w.handleJobError(job, err)
	} else {
		w.handleJobSuccess(job)
	}
}

// processVirusScan performs virus scanning
func (w *FileWorker) processVirusScan(job *domain.FileProcessingJob) error {
	// Get file
	file, err := w.fileRepo.GetByID(w.ctx, job.FileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Mock virus scan - in real implementation, this would call actual virus scanner
	log.Printf("Mock virus scan for file: %s", file.OriginalName)

	// Update file with scan results
	file.VirusScanned = true
	scanResult := "clean"
	file.VirusScanResult = &scanResult
	file.Status = domain.FileStatusAvailable

	return w.fileRepo.Update(w.ctx, file)
}

// processImageResize resizes an image
func (w *FileWorker) processImageResize(job *domain.FileProcessingJob) error {
	type ResizeParams struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	var params ResizeParams
	if err := json.Unmarshal(job.Parameters, &params); err != nil {
		return fmt.Errorf("failed to parse resize parameters: %w", err)
	}

	// Get file
	file, err := w.fileRepo.GetByID(w.ctx, job.FileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Mock image resize - in real implementation, this would call actual image processor
	log.Printf("Mock image resize for file: %s to %dx%d", file.OriginalName, params.Width, params.Height)

	// Update file metadata
	if file.Metadata == nil {
		file.Metadata = make(map[string]interface{})
	}
	file.Metadata["resized"] = true
	file.Metadata["resize_params"] = params

	return w.fileRepo.Update(w.ctx, file)
}

// processImageCrop crops an image
func (w *FileWorker) processImageCrop(job *domain.FileProcessingJob) error {
	type CropParams struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	var params CropParams
	if err := json.Unmarshal(job.Parameters, &params); err != nil {
		return fmt.Errorf("failed to parse crop parameters: %w", err)
	}

	// Get file
	file, err := w.fileRepo.GetByID(w.ctx, job.FileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Mock image crop - in real implementation, this would call actual image processor
	log.Printf("Mock image crop for file: %s", file.OriginalName)

	// Update file metadata
	if file.Metadata == nil {
		file.Metadata = make(map[string]interface{})
	}
	file.Metadata["cropped"] = true
	file.Metadata["crop_params"] = params

	return w.fileRepo.Update(w.ctx, file)
}

// processThumbnailGeneration generates thumbnails
func (w *FileWorker) processThumbnailGeneration(job *domain.FileProcessingJob) error {
	// Get file
	file, err := w.fileRepo.GetByID(w.ctx, job.FileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Mock thumbnail generation - in real implementation, this would call actual image processor
	log.Printf("Mock thumbnail generation for file: %s", file.OriginalName)

	// Update file metadata
	if file.Metadata == nil {
		file.Metadata = make(map[string]interface{})
	}
	file.Metadata["thumbnail_generated"] = true
	file.Metadata["thumbnail_path"] = fmt.Sprintf("%s_thumbnail", file.StoragePath)

	return w.fileRepo.Update(w.ctx, file)
}

// processMetadataExtraction extracts file metadata
func (w *FileWorker) processMetadataExtraction(job *domain.FileProcessingJob) error {
	// Get file
	file, err := w.fileRepo.GetByID(w.ctx, job.FileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Mock metadata extraction - in real implementation, this would extract actual metadata
	log.Printf("Mock metadata extraction for file: %s", file.OriginalName)

	// Update file metadata
	if file.Metadata == nil {
		file.Metadata = make(map[string]interface{})
	}
	file.Metadata["metadata_extracted"] = true
	file.Metadata["extraction_time"] = time.Now()

	return w.fileRepo.Update(w.ctx, file)
}

// handleJobSuccess handles successful job completion
func (w *FileWorker) handleJobSuccess(job *domain.FileProcessingJob) {
	job.Status = domain.JobStatusCompleted
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Progress = 100

	if err := w.jobRepo.Update(w.ctx, job); err != nil {
		log.Printf("Error updating successful job: %v", err)
	}

	log.Printf("Job completed successfully: %s", job.ID)
}

// handleJobError handles job errors and retries
func (w *FileWorker) handleJobError(job *domain.FileProcessingJob, err error) {
	job.RetryCount++
	job.LastError = err.Error()

	if job.RetryCount >= w.maxRetries {
		job.Status = domain.JobStatusFailed
		failedAt := time.Now()
		job.FailedAt = &failedAt
		log.Printf("Job failed after %d retries: %s - %v", w.maxRetries, job.ID, err)
	} else {
		job.Status = domain.JobStatusPending
		// Schedule retry
		go func() {
			time.Sleep(w.retryDelay * time.Duration(job.RetryCount))
			w.SubmitJob(job)
		}()
		log.Printf("Job failed, will retry (%d/%d): %s - %v", job.RetryCount, w.maxRetries, job.ID, err)
	}

	if updateErr := w.jobRepo.Update(w.ctx, job); updateErr != nil {
		log.Printf("Error updating failed job: %v", updateErr)
	}
}
