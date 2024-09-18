from functools import cmp_to_key

PAWN = 0
KNIGHT = 1
BISHOP = 2
ROOK = 3
QUEEN = 4
KING = 5

capture_values = [100, 300, 301, 500, 900, 1000]
string_values = ["P", "N", "B", "R", "Q", "K"]

def compare_func(a, b):
    if a[0] == b[0]:
        return b[1]-a[1]
    return a[0] - b[0]
    
def main():
    attack_list = []
    for attacker in range(KING+1):
        for victim in range(KING):
            attack_list.append([capture_values[victim] - capture_values[attacker], capture_values[attacker], string_values[attacker] + "x" +string_values[victim] + "_mvv"])

    # for attacker in range(KING+1):
    #     for victim in range(KING):
    #         attack_list.append([capture_values[victim], capture_values[attacker], string_values[attacker] + "x" +string_values[victim] + "_blind"])
    sort_list = sorted(attack_list, key=cmp_to_key(compare_func))
    
    for i, elm in enumerate(sort_list):
        print(i, elm)
    
main()