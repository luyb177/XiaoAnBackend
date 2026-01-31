package tasks

type TaskPrefix string

const (
	ArticleRelationTaskPrefix      TaskPrefix = "article_relation_task"
	VideoRelationTaskPrefix        TaskPrefix = "video_relation_task"
	PodcastRelationTaskPrefix      TaskPrefix = "podcast_relation_task"
	ComicRelationTaskPrefix        TaskPrefix = "comic_relation_task"
	ComicChapterRelationTaskPrefix TaskPrefix = "comic_chapter_relation_task"
	CommentRelationTaskPrefix      TaskPrefix = "comment_relation_task"
)
