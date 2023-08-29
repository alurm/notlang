#include <stdlib.h>
#include <assert.h>
#include "h"

void add_token(tokens *tokens1, token token1) {
	int i;

	if (tokens1->length >= tokens1->capacity) {
		tokens tokens2;
		tokens2.length = tokens1->length;
		tokens2.capacity = tokens1->capacity * 2 + 1;
		tokens2.value = malloc(sizeof (token) * tokens2.capacity);
		assert(tokens2.value != 0);

		for (i = 0; i < tokens1->length; i++) {
			tokens2.value[i] = tokens1->value[i];
		}
		free(tokens1->value);
		*tokens1 = tokens2;
	}

	tokens1->value[tokens1->length] = token1;
	tokens1->length++;
}

tokens tokenize() {
	tokens tokens1 = { 0 };
	token token1;
	token1.tag = token_tag_separator;
	add_token(&tokens1, token1);
	add_token(&tokens1, token1);
	//add_token(&tokens1, token1);
}

int main(void) {
	tokenize();
}
