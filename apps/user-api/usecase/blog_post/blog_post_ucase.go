package blog_post

import (
	"context"
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	udoamin "github.com/illuminarean/energyshares-user-api/domain"
)

type blogPostUsecase struct {
	blogPostRepo cdomain.BlogPostRepository
}

func New(repo cdomain.BlogPostRepository) udoamin.BlogPostUsecase {
	return &blogPostUsecase{
		blogPostRepo: repo,
	}
}

func (u blogPostUsecase) GetAll(ctx context.Context, request udoamin.BlogPostGetAllRequest) (int64, udoamin.BlogPostGetAllResponse, error) {
	if request.Limit == 0 {
		DISPLAY_LIMIT := 6 // default limit
		request.Limit = DISPLAY_LIMIT
		request.Offset = 0
	}

	totalCnt, items, err := u.blogPostRepo.GetAll(ctx, request.Offset, request.Limit)
	if err != nil {
		return 0, udoamin.BlogPostGetAllResponse{}, err
	}

	response := make([]udoamin.BlogPostResponse, 0, len(items))
	for _, item := range items {
		response = append(response, udoamin.BlogPostResponse{
			Id:       item.Id,
			Title:    item.Title,
			Link:     item.Link,
			Data:     item.Data,
			Category: item.Category,
			Img:      item.SourceUrl,
			Date:     item.PublishedAt.Time().Format(time.DateOnly),
		})
	}

	return totalCnt, udoamin.BlogPostGetAllResponse{
		TotalCount: totalCnt,
		Items:      response,
	}, nil
}
