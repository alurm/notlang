typedef struct string {
	int length;
	int capacity;
	char *buffer;
} string;

typedef struct value {
	enum { type_string, type_list } type;
	union { string string; struct list *list; };
} value;

typedef struct list { value value; struct list *next; } list;

typedef struct token {
	enum {
		token_open,
		token_close,
		token_string,
	} type;
	union {
		string string;
	};
} token;

typedef struct token_list {
	token value;
	struct token_list *next;
} token_list;
