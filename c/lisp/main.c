#include "main.h"

#include <stdlib.h>
#include <stdio.h>
#include <assert.h>
#include <stdbool.h>
#include <string.h>

void *mycalloc(size_t size) { void *ptr = calloc(1, size); assert(ptr); return ptr; }

void push_char(string *s, char c) {
	if (s->length == s->capacity) {
		int new_capacity = s->capacity * 2 + 1;
		char *new_buffer = mycalloc(new_capacity);
		for (int i = 0; i < s->length; i++) new_buffer[i] = s->buffer[i];
		free(s->buffer);
		s->buffer = new_buffer;
	}
	s->buffer[s->length] = c;
	s->length++;
}

void push_token(token_list **ptr, token t) {
	token_list *new = mycalloc(sizeof *new);
	*new = (token_list){
		next: *ptr,
		value: t,
	};
	*ptr = new;
}

void print_token(token t) {
	switch (t.type) {
	case token_open:
		printf("open\n");
		break;
	case token_close:
		printf("close\n");
		break;
	case token_string:
		printf("string: ");
		for (int i = 0; i < t.string.length; i++)
			printf("%c", t.string.buffer[i]);
		printf("\n");
		break;
	default:
		assert(0);
	}
}

void print_tokens(token_list *list) {
	for (; list != 0; list = list->next) {
		print_token(list->value);
	}
}

token_list *tokenize(char *source) {
	token_list *result = 0;
	for (;;) {
		switch (*source) {
		case '(':
			push_token(&result, (token){ type: token_open });
			source++;
			break;
		case ')':
			push_token(&result, (token){ type: token_close });
			source++;
			break;
		case ' ':
		case '\t':
		case '\n':
			source++;
			break;
		case 0:
			return result;
		default:
			string s = {0};
			for (;;) {
				char c = *source;
				switch (c) {
				case '(':
				case ')':
				case ' ':
				case '\t':
				case '\n':
					push_token(&result, (token){ type: token_string, string: s });
					goto string_is_read;
				case 0:
					push_token(&result, (token){ type: token_string, string: s });
					return result;
				default:
					push_char(&s, c);
					source++;
				}
			}
		string_is_read:
		}
	}
}

token_list *reverse_tokens(token_list *in) {
	token_list *out = 0;
	for (; in != 0; ) {
		push_token(&out, in->value);
		token_list *next = in->next;
		free(in);
		in = next;
	}
	return out;
}

void push_value(list **ptr, value v) {
	list *new = mycalloc(sizeof *new);
	*new = (list){
		value: v,
		next: *ptr,
	};
	*ptr = new;
}

struct parse {
	value value;
	token_list *next;
};

struct parse parse_list(token_list *in) {
	list *out = 0;
	for (; in != 0;) {
		switch (in->value.type) {
		case token_close:
			return (struct parse){
				value: (value){ type: type_list, list: out },
				next: in->next,
			};
		case token_open:
			struct parse p = parse_list(in->next);
			in = p.next;
			push_value(&out, p.value);
			break;
		case token_string:
			push_value(&out, (value){
				type: type_string,
				string: in->value.string,
			});
			in = in->next;
			break;
		default:
			assert(0);
		}
	}
	assert(0);
}

void free_list(list *in) {
	for (; in != 0; in = in->next) {
		switch (in->value.type) {
		case type_string:
			break;
		case type_list:
			free_list(in->value.list);
			break;
		default:
			assert(0);
		}
	}
}

list *reverse_list(list *in) {
	list *out = 0;
	for (; in != 0; in = in->next) {
		switch (in->value.type) {
		case type_string:
			push_value(&out, in->value);
			break;
		case type_list:
			list *under = reverse_list(in->value.list);
			push_value(
				&out,
				(value){
					type: type_list,
					list: under,
				}
			);
			break;
		default:
			assert(0);
		}
	}
	free_list(in);
	return out;
}

struct parse parse(token_list *in) {
	struct parse out;
	for (;;) {
		switch (in->value.type) {
		case token_close:
			assert(0);
		case token_open:
			struct parse p = parse_list(in->next);
			p.value.list = reverse_list(p.value.list);
			return p;
		case token_string:
			return (struct parse){
				value: (value){ type: type_string, string: in->value.string },
				next: in->next,
			};
		default:
			assert(0);
		}
	}
}

value parse_one(token_list *in) {
	struct parse p = parse(in);
	return p.value;
}

void print_string(string s) {
	for (int i = 0; i < s.length; i++)
		printf("%c", s.buffer[i]);
}

void tab(int depth) { for (int i = 0; i < depth; i++) printf("\t"); }

void print_value_depth(value v, int depth) {
	switch (v.type) {
	case type_string:
		tab(depth);
		print_string(v.string);
		printf("\n");
		break;
	case type_list:
		list *current = v.list;
		tab(depth); printf("(\n");
		for (; current != 0; current = current->next) {
			print_value_depth(current->value, depth + 1);
		}
		tab(depth); printf(")\n");
		break;
	default:
		assert(0);
	};
}

void print_value(value v) { print_value_depth(v, 0); }

void tests(void) {
	print_tokens(tokenize("(((("));
	{
		string s = {0};
		push_char(&s, 'a');
		push_char(&s, 'b');
		push_char(&s, 'c');
		push_char(&s, 0);
		printf("%s\n", s.buffer);
	}
	{
		token_list *tokens = reverse_tokens(tokenize("(foo () (bar baz))"));
		int first = true;
		for (; tokens != 0;) {
			if (!first) printf("\n");
			else first = false;

			struct parse p = parse(tokens);
			tokens = p.next;
			print_value(p.value);
		}
	}
}

typedef struct environment {
	string key;
	value value;
	struct environment *up;
} environment;



bool are_equal(string l, string r) {
	if (l.length != r.length) {
		return false;
	}
	int length = l.length;
	for (int i = 0; i < length; i++) {
		if (l.buffer[i] != r.buffer[i])
			return false;
	}
	return true;
}

value evaluate(value form, environment *e) {
	switch (form.type) {
	case type_string:
		// Shells usually do this.
		return form;
		// Lisps usually do this.
		for (; e != 0; e = e->up) {
			if (are_equal(e->key, form.string)) {
				return e->value;
			}
		}
		assert(0);
	case type_list:
		list *l = form.list;
		assert(l);
		value head = l->value;
		list *tail = l->next;
		assert(0);
	}
}

value apply(list *l) {

}

void argv(int argc, char **argv) {
	assert(argc == 2);
	token_list *tokens = reverse_tokens(tokenize(argv[1]));
	int first = true;
	for (; tokens != 0;) {
		if (!first) printf("\n");
		else first = false;

		struct parse p = parse(tokens);
		tokens = p.next;
		print_value(p.value);
	}
}

char *clone(char *s) {
	int length = 0;
	for (char *s2 = s; *s2; s2++)
		length++;
	char *new = mycalloc(length + 1);
	for (int i = 0; s[i]; i++)
		new[i] = s[i];
	new[length] = 0;
	return new;
}

string chars2string(char *s) {
	int length = strlen(s);
	return (string){
		length: length,
		capacity: length,
		buffer: clone(s),
	};
}

int main(int argc, char **argv) {
	print_value(evaluate(parse_one(tokenize("hello")), &(environment){
		key: chars2string("hello"),
		value: (value){ type_string, chars2string("world") },
		up: 0,
	}));
}
