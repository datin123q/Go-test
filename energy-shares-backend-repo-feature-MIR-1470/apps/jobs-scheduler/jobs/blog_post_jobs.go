package jobs

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/ghost"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type BlogPostJobs struct{}

type BlogApiServiceResponse struct {
}

func (BlogPostJobs) BlogPostRequest(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_BLOG_POST_REQUEST,
		JobName: jobNameBlogPostRequest,
	}, blogPostRequest)
}

func blogPostRequest(ctx context.Context) error {
	posts, err := ghost.RequestTopPosts(serviceutils.GetBlogApiServiceUrl(), serviceutils.GetBlogApiServiceKey())
	if err != nil {
		return errors.Wrap(err, "failed to request blog api service")
	}

	logrus.WithFields(logrus.Fields{
		"posts": posts,
	}).Debug("BlogPostJobs.BlogPostRequest")

	for _, post := range posts.Posts {

		categoryName := ""
		if len(post.FeatureImage) == 0 {
			continue
		}

		createdAt, err := jsonutils.ParseJSONTimeJST(time.RFC3339, post.CreatedAt)
		if err != nil {
			return errors.Wrap(err, "failed to parse write date")
		}

		publishedAt, err := jsonutils.ParseJSONTimeJST(time.RFC3339, post.PublishedAt)
		if err != nil {
			return errors.Wrap(err, "failed to parse write date")
		}

		blogPost := domain.BlogPost{
			Id:          post.Id,
			Title:       post.Title,
			Link:        post.URL,
			Data:        post.HTML,
			Category:    categoryName,
			SourceUrl:   post.FeatureImage,
			WriteAt:     createdAt,
			PublishedAt: publishedAt,
		}
		exist, err := blogPost.IsExist(ctx, post.Id)
		if err != nil {
			return errors.Wrap(err, "failed to check blog post exist")
		}
		if !exist {
			_, err = blogPost.Create(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to create blog post")
			}
		}
		err = blogPost.Update(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to update blog post")
		}

	}

	return nil
}
