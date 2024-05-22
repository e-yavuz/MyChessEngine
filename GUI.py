import pygame, sys
from pygame.locals import *
from typing import Dict

# --- constants --- (UPPER_CASE names)

BACKGROUND = (256-0xBA,256-0xA3,256-0xA9)
BROWN_TILE = (120, 70, 45)
GREY_TILE = (130, 130, 130)
x = 1200
y = 980
TILE_SIZE = (int) (x*0.8)/8 if (int) (x*0.8)/8 < (int) (y*0.8)/8 else (int) (y*0.8)/8

X_OFFSET = (int) (x-(TILE_SIZE*8))/2
Y_OFFSET = (int) (y-(TILE_SIZE*8))/2

PIECE_SIZE = (TILE_SIZE, TILE_SIZE)

POSITION_TO_PIECE : Dict[int, pygame.Rect] = {}
PIECE_TO_IMAGE = {}

CHESS_IMAGES = pygame.image.load("Images/chessImages.png")
selectedRectangle = None


                

# --- functions --- (lower_case names)

def clip(surface, x, y, x_size, y_size): #Get a part of the image
    handle_surface = surface.copy() #Sprite that will get process later
    clipRect = pygame.Rect(x,y,x_size,y_size) #Part of the image
    handle_surface.set_clip(clipRect) #Clip
    image = surface.subsurface(handle_surface.get_clip()) #Get subsurface
    return image.copy() #Return

def draw_FEN(FEN : str, DISPLAY : pygame.Surface, PIECE_TO_IMAGE : dict):
    if(FEN is None):
        raise ValueError("FEN ", FEN, " String cannot be None")
    
    position = 0
    for char in FEN:
        file = position%8
        rank = int(position/8)
        
        if char == "/":
            continue
        
        elif char.isalpha():
            if (file+rank)%2 == 0:
                    COLOR = BROWN_TILE
            else:
                COLOR = GREY_TILE
            pygame.draw.rect(DISPLAY,COLOR,
                             (
                                X_OFFSET+(TILE_SIZE*file), 
                                Y_OFFSET+(TILE_SIZE*rank),
                                TILE_SIZE,TILE_SIZE))
            if PIECE_TO_IMAGE.get(char):
                POSITION_TO_PIECE[position] = DISPLAY.blit(
                                PIECE_TO_IMAGE[char], 
                                ((X_OFFSET+(TILE_SIZE*file),
                                Y_OFFSET+(TILE_SIZE*rank))))
            position+=1
            
        elif char.isdigit():
            for _ in range(0, int(char)):
                file = position%8
                rank = int(position/8)
                if (file+rank)%2 == 0:
                    COLOR = BROWN_TILE
                else:
                    COLOR = GREY_TILE
                pygame.draw.rect(DISPLAY,COLOR,(X_OFFSET+(TILE_SIZE*file), Y_OFFSET+(TILE_SIZE*rank),TILE_SIZE,TILE_SIZE))
                position+=1

# --- main ---

def main():
    pygame.init()
    
    PIECE_TO_IMAGE = {
        "K": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*0/6),0,134,134), PIECE_SIZE),
        "Q": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*1/6),0,134,134), PIECE_SIZE),
        "B": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*2/6),0,134,134), PIECE_SIZE),
        "N": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*3/6),0,134,134), PIECE_SIZE),
        "R": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*4/6),0,134,134), PIECE_SIZE),
        "P": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*5/6),0,134,134), PIECE_SIZE),
        "k": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*0/6),134,134,134), PIECE_SIZE),
        "q": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*1/6),134,134,134), PIECE_SIZE),
        "b": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*2/6),134,134,134), PIECE_SIZE),
        "n": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*3/6),134,134,134), PIECE_SIZE),
        "r": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*4/6),134,134,134), PIECE_SIZE),
        "p": 
        pygame.transform.scale(clip(CHESS_IMAGES, (800*5/6),134,134,134), PIECE_SIZE)
    }

    
    DISPLAY = pygame.display.set_mode((x,y),0,32)

    DISPLAY.fill(BACKGROUND)
    
    
    # Starting FEN: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
    draw_FEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR", DISPLAY, PIECE_TO_IMAGE)
    
    selectedRectangle = None
    offset_x = 0
    offset_y = 0

    while True:
        for event in pygame.event.get():
            if event.type==QUIT:
                pygame.quit()
                sys.exit()
            
            elif event.type == pygame.MOUSEBUTTONDOWN:
                if event.button == 1 :
                    for key in POSITION_TO_PIECE:
                        piece = POSITION_TO_PIECE[key]
                        if piece.collidepoint(event.pos):
                            selectedRectangle = POSITION_TO_PIECE[key]
                            mouse_x, mouse_y = event.pos
                            offset_x = selectedRectangle.x - mouse_x
                            offset_y = selectedRectangle.y - mouse_y
                            break

            elif event.type == pygame.MOUSEBUTTONUP:
                if event.button == 1:  
                    selectedRectangle = None
                    print("Dropped piece")

            elif event.type == pygame.MOUSEMOTION:
                if selectedRectangle is not None:
                    mouse_x, mouse_y = event.pos
                    selectedRectangle.x = mouse_x + offset_x
                    selectedRectangle.y = mouse_y + offset_y
        pygame.display.flip()

main()